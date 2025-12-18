package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"playplus_platform/internal/config"
)

func init() {
	// Seed random number generator for jitter
	rand.Seed(time.Now().UnixNano())
}

const (
	maxRetries     = 3
	baseRetryDelay = 500 * time.Millisecond
	maxRetryDelay  = 5 * time.Second
)

// VModel API version IDs
const (
	VModelVideoFaceDetectVersion    = "fa9317a2ad086f7633f4f9b38f35c82495b6c5f38fa2afbe32d9d9df8620b389"
	VModelVideoMultiFaceSwapVersion = "8e960283784c5b58e5f67236757c40bb6796c85e3c733d060342bdf62f9f0c64"
)

// VModelClient handles all VModel API interactions
type VModelClient struct {
	httpClient *http.Client
	cfg        *config.Config
}

var (
	vmodelClient *VModelClient
	vmodelOnce   sync.Once
)

// GetVModelClient returns the singleton VModel client
func GetVModelClient() *VModelClient {
	vmodelOnce.Do(func() {
		vmodelClient = &VModelClient{
			httpClient: &http.Client{Timeout: 120 * time.Second},
			cfg:        config.Get(),
		}
	})
	return vmodelClient
}

// --- Request/Response Types ---

type vmodelCreateTaskRequest struct {
	Version string                 `json:"version"`
	Input   map[string]interface{} `json:"input"`
}

// vmodelAPIResponse is the wrapper response from VModel API
type vmodelAPIResponse struct {
	Code    int             `json:"code"`
	Result  json.RawMessage `json:"result"`
	Message json.RawMessage `json:"message"`
}

// vmodelCreateTaskResult is the result from create task API
type vmodelCreateTaskResult struct {
	TaskID   string `json:"task_id"`
	TaskCost int    `json:"task_cost"`
}

// vmodelTaskResult is the result from get task API
type vmodelTaskResult struct {
	TaskID      string          `json:"task_id"`
	UserID      int             `json:"user_id"`
	Version     string          `json:"version"`
	Error       json.RawMessage `json:"error"`
	TotalTime   float64         `json:"total_time"`
	PredictTime float64         `json:"predict_time"`
	Logs        *string         `json:"logs"`
	Output      json.RawMessage `json:"output"`
	Status      string          `json:"status"` // "starting", "processing", "succeeded", "failed"
	CreateAt    int64           `json:"create_at"`
	CompletedAt *int64          `json:"completed_at"`
}

// VModelDetectedFace represents a detected face from VModel
type VModelDetectedFace struct {
	ID   int    `json:"id"`
	Link string `json:"link"` // Thumbnail URL
}

type vmodelDetectOutput struct {
	ID         string               `json:"id"` // detect_id
	Status     string               `json:"status"`
	Type       string               `json:"type"`
	Faces      []VModelDetectedFace `json:"faces"`
	StartedAt  interface{}          `json:"started_at"`  // Can be int64 or string
	FinishedAt interface{}          `json:"finished_at"` // Can be int64 or string
	Error      *string              `json:"error"`
}

// VModelDetectResult is the result of face detection
type VModelDetectResult struct {
	DetectID string               `json:"detect_id"`
	Faces    []VModelDetectedFace `json:"faces"`
}

// VModelFaceSwapPair represents a face swap mapping
type VModelFaceSwapPair struct {
	FaceID int    `json:"face_id"`
	Target string `json:"target"` // Source image URL (new face)
}

// VModelSwapTaskResult is the result of a swap task
type VModelSwapTaskResult struct {
	TaskID    string `json:"task_id"`
	Status    string `json:"status"` // queuing, processing, completed, failed
	ResultURL string `json:"result_url,omitempty"`
	Error     string `json:"error,omitempty"`
}

// VModelDetectTaskResult is the result of a detect task creation
type VModelDetectTaskResult struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"` // queuing, processing, completed, failed
}

// VModelDetectStatusResult is the result of checking detect task status
type VModelDetectStatusResult struct {
	TaskID   string               `json:"task_id"`
	Status   string               `json:"status"` // queuing, processing, completed, failed
	DetectID string               `json:"detect_id,omitempty"`
	Faces    []VModelDetectedFace `json:"faces,omitempty"`
	Error    string               `json:"error,omitempty"`
}

// --- API Methods ---

// doRequest makes an authenticated request to VModel API
// GET requests are retried on transient failures, POST requests are not (not idempotent)
func (c *VModelClient) doRequest(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error) {
	var jsonBody []byte
	var err error
	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
	}

	url := fmt.Sprintf("%s%s", c.cfg.VModelBaseURL, endpoint)

	// Only retry GET requests (idempotent)
	retryable := method == "GET"
	maxAttempts := 1
	if retryable {
		maxAttempts = maxRetries + 1
	}

	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Wait before retry (skip first attempt)
		if attempt > 0 {
			delay := time.Duration(1<<uint(attempt-1)) * baseRetryDelay
			if delay > maxRetryDelay {
				delay = maxRetryDelay
			}
			// Add jitter (0-25% of delay)
			jitter := time.Duration(rand.Int63n(int64(delay / 4)))
			delay += jitter

			log.Printf("[WARN] VModel API retry %d/%d for %s %s after %v", attempt, maxRetries, method, endpoint, delay)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		// Create request with fresh body reader
		var reqBody io.Reader
		if jsonBody != nil {
			reqBody = bytes.NewReader(jsonBody)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.cfg.VModelAPIToken)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			// Only short-circuit if caller's context is done
			if ctx.Err() != nil {
				return nil, lastErr
			}
			if retryable {
				continue // Retry
			}
			return nil, lastErr
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("read response: %w", err)
			if retryable {
				continue // Retry
			}
			return nil, lastErr
		}

		// Check for retryable HTTP status codes (5xx, 429)
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
			if retryable {
				continue // Retry
			}
			return nil, lastErr
		}

		// Parse the wrapper response
		var apiResp vmodelAPIResponse
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			return nil, fmt.Errorf("decode API response: %w (body: %s)", err, string(respBody))
		}

		// Check for API-level errors (not retryable - business logic errors)
		if apiResp.Code != 200 {
			return nil, fmt.Errorf("API error (code %d): %s", apiResp.Code, string(apiResp.Message))
		}

		return apiResp.Result, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// DetectFaces detects faces in an image/video URL
func (c *VModelClient) DetectFaces(ctx context.Context, mediaURL string) (*VModelDetectResult, error) {
	reqBody := vmodelCreateTaskRequest{
		Version: VModelVideoFaceDetectVersion,
		Input: map[string]interface{}{
			"source": mediaURL,
		},
	}

	resultBytes, err := c.doRequest(ctx, "POST", "/api/tasks/v1/create", reqBody)
	if err != nil {
		return nil, fmt.Errorf("create detect task: %w", err)
	}

	var createResult vmodelCreateTaskResult
	if err := json.Unmarshal(resultBytes, &createResult); err != nil {
		return nil, fmt.Errorf("decode create result: %w (body: %s)", err, string(resultBytes))
	}

	// Poll for completion (max 120 seconds for video detection)
	taskResult, err := c.waitForTask(ctx, createResult.TaskID, 120*time.Second)
	if err != nil {
		return nil, err
	}

	if taskResult.Status == "failed" {
		errMsg := parseVModelError(taskResult.Error)
		if errMsg == "" {
			errMsg = "unknown error"
		}
		return nil, fmt.Errorf("detection failed: %s", errMsg)
	}

	// Parse output - it's an array of detect outputs
	var outputs []vmodelDetectOutput
	if err := json.Unmarshal(taskResult.Output, &outputs); err != nil {
		return nil, fmt.Errorf("decode output: %w (output: %s)", err, string(taskResult.Output))
	}

	if len(outputs) == 0 {
		return nil, fmt.Errorf("no detection output")
	}

	// Aggregate faces from ALL outputs (not just the first one)
	var allFaces []VModelDetectedFace
	var detectID string

	for _, output := range outputs {
		if output.Status == "succeed" {
			if detectID == "" {
				detectID = output.ID // Use first successful detect_id
			}
			allFaces = append(allFaces, output.Faces...)
		}
	}

	if detectID == "" {
		// No successful detection
		errMsg := "detection not successful"
		if outputs[0].Error != nil {
			errMsg = *outputs[0].Error
		}
		return nil, fmt.Errorf(errMsg)
	}

	return &VModelDetectResult{
		DetectID: detectID,
		Faces:    allFaces,
	}, nil
}

// CreateDetectTask creates a face detection task and returns immediately (async)
func (c *VModelClient) CreateDetectTask(ctx context.Context, mediaURL string) (*VModelDetectTaskResult, error) {
	reqBody := vmodelCreateTaskRequest{
		Version: VModelVideoFaceDetectVersion,
		Input: map[string]interface{}{
			"source": mediaURL,
		},
	}

	resultBytes, err := c.doRequest(ctx, "POST", "/api/tasks/v1/create", reqBody)
	if err != nil {
		return nil, fmt.Errorf("create detect task: %w", err)
	}

	var createResult vmodelCreateTaskResult
	if err := json.Unmarshal(resultBytes, &createResult); err != nil {
		return nil, fmt.Errorf("decode create result: %w (body: %s)", err, string(resultBytes))
	}

	return &VModelDetectTaskResult{
		TaskID: createResult.TaskID,
		Status: "queuing",
	}, nil
}

// GetDetectTaskStatus gets the status of a face detection task
// Error handling:
// - queuing/processing: returns result, nil error (caller keeps polling)
// - succeeded: returns result, nil error (success)
// - failed: returns result, error (actual failure - caller should stop polling)
// - HTTP/decode errors: returns nil, error
func (c *VModelClient) GetDetectTaskStatus(ctx context.Context, taskID string) (*VModelDetectStatusResult, error) {
	endpoint := fmt.Sprintf("/api/tasks/v1/get/%s", taskID)
	resultBytes, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("get detect task: %w", err)
	}

	var taskResult vmodelTaskResult
	if err := json.Unmarshal(resultBytes, &taskResult); err != nil {
		return nil, fmt.Errorf("decode task result: %w", err)
	}

	status := c.mapStatus(taskResult.Status)
	result := &VModelDetectStatusResult{
		TaskID: taskID,
		Status: status,
	}

	// For queuing/processing, return status without error (caller keeps polling)
	if taskResult.Status == "starting" || taskResult.Status == "processing" {
		return result, nil
	}

	// For failed status, return both result and error
	if taskResult.Status == "failed" {
		result.Error = parseVModelError(taskResult.Error)
		if result.Error == "" {
			result.Error = "detection failed"
		}
		return result, fmt.Errorf("detection failed: %s", result.Error)
	}

	// For succeeded status, parse the output
	if taskResult.Status == "succeeded" {
		var outputs []vmodelDetectOutput
		if err := json.Unmarshal(taskResult.Output, &outputs); err != nil {
			return nil, fmt.Errorf("decode output: %w", err)
		}

		// Aggregate faces from all outputs
		for _, output := range outputs {
			if output.Status == "succeed" {
				if result.DetectID == "" {
					result.DetectID = output.ID
				}
				result.Faces = append(result.Faces, output.Faces...)
			}
		}

		if result.DetectID == "" {
			// No faces detected is a valid completed state, not a failure
			result.Status = "completed"
			result.Faces = []VModelDetectedFace{} // Empty faces array
			// Don't return error - this is a successful completion with no faces
		}
	}

	return result, nil
}

// CreateSwapTask creates a video face swap task
func (c *VModelClient) CreateSwapTask(ctx context.Context, detectID string, faceSwaps []VModelFaceSwapPair, faceEnhance bool) (*VModelSwapTaskResult, error) {
	// Build face_map JSON string
	faceMapJSON, err := json.Marshal(faceSwaps)
	if err != nil {
		return nil, fmt.Errorf("marshal face_map: %w", err)
	}

	reqBody := vmodelCreateTaskRequest{
		Version: VModelVideoMultiFaceSwapVersion,
		Input: map[string]interface{}{
			"detect_id":    detectID,
			"face_map":     string(faceMapJSON),
			"face_enhance": faceEnhance,
		},
	}

	resultBytes, err := c.doRequest(ctx, "POST", "/api/tasks/v1/create", reqBody)
	if err != nil {
		return nil, fmt.Errorf("create swap task: %w", err)
	}

	var createResult vmodelCreateTaskResult
	if err := json.Unmarshal(resultBytes, &createResult); err != nil {
		return nil, fmt.Errorf("decode create result: %w (body: %s)", err, string(resultBytes))
	}

	return &VModelSwapTaskResult{
		TaskID: createResult.TaskID,
		Status: "queuing",
	}, nil
}

// GetTaskStatus retrieves the status of a task
func (c *VModelClient) GetTaskStatus(ctx context.Context, taskID string) (*VModelSwapTaskResult, error) {
	// Note: endpoint is /api/tasks/v1/get/{task_id}
	endpoint := fmt.Sprintf("/api/tasks/v1/get/%s", taskID)
	resultBytes, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("get task status: %w", err)
	}

	var taskResult vmodelTaskResult
	if err := json.Unmarshal(resultBytes, &taskResult); err != nil {
		return nil, fmt.Errorf("decode task result: %w (body: %s)", err, string(resultBytes))
	}

	result := &VModelSwapTaskResult{
		TaskID: taskResult.TaskID,
		Status: c.mapStatus(taskResult.Status),
	}

	if errMsg := parseVModelError(taskResult.Error); errMsg != "" {
		result.Error = errMsg
	}

	// Parse output for result URL if completed
	if taskResult.Status == "succeeded" && len(taskResult.Output) > 0 {
		// Output is an array of URLs
		var outputArr []string
		if err := json.Unmarshal(taskResult.Output, &outputArr); err != nil {
			return nil, fmt.Errorf("failed to parse output URLs: %w (output: %s)", err, string(taskResult.Output))
		}
		if len(outputArr) > 0 {
			result.ResultURL = outputArr[0]
		}
	}

	return result, nil
}

// waitForTask polls for task completion
func (c *VModelClient) waitForTask(ctx context.Context, taskID string, timeout time.Duration) (*vmodelTaskResult, error) {
	deadline := time.Now().Add(timeout)
	pollInterval := 2 * time.Second

	for time.Now().Before(deadline) {
		endpoint := fmt.Sprintf("/api/tasks/v1/get/%s", taskID)
		resultBytes, err := c.doRequest(ctx, "GET", endpoint, nil)
		if err != nil {
			return nil, err
		}

		var taskResult vmodelTaskResult
		if err := json.Unmarshal(resultBytes, &taskResult); err != nil {
			return nil, fmt.Errorf("decode task result: %w", err)
		}

		if taskResult.Status == "succeeded" || taskResult.Status == "failed" {
			return &taskResult, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollInterval):
			// Continue polling
		}
	}

	return nil, fmt.Errorf("task timeout after %v", timeout)
}

// parseVModelError handles error fields that may be strings or objects.
func parseVModelError(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}

	var errStr string
	if err := json.Unmarshal(raw, &errStr); err == nil {
		return errStr
	}

	return string(raw)
}

// mapStatus maps VModel status to our standard status
func (c *VModelClient) mapStatus(status string) string {
	switch status {
	case "starting":
		return "queuing"
	case "processing":
		return "processing"
	case "succeeded":
		return "completed"
	case "failed":
		return "failed"
	default:
		return status
	}
}

// VModelStatusToString converts status for API response
func VModelStatusToString(status string) string {
	return status // Already a string
}

// GetCredits retrieves the current credits balance
func (c *VModelClient) GetCredits(ctx context.Context) (float64, error) {
	resultBytes, err := c.doRequest(ctx, "POST", "/api/users/v1/account/credits/left", map[string]interface{}{})
	if err != nil {
		return 0, fmt.Errorf("get credits: %w", err)
	}

	var credits float64
	if err := json.Unmarshal(resultBytes, &credits); err != nil {
		return 0, fmt.Errorf("decode credits: %w (body: %s)", err, string(resultBytes))
	}

	return credits, nil
}
