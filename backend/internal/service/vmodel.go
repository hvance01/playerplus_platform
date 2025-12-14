package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"playplus_platform/internal/config"
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
	Error       *string         `json:"error"`
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

// --- API Methods ---

// doRequest makes an authenticated request to VModel API
func (c *VModelClient) doRequest(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	url := fmt.Sprintf("%s%s", c.cfg.VModelBaseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.VModelAPIToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Parse the wrapper response
	var apiResp vmodelAPIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("decode API response: %w (body: %s)", err, string(respBody))
	}

	// Check for API-level errors
	if apiResp.Code != 200 {
		return nil, fmt.Errorf("API error (code %d): %s", apiResp.Code, string(apiResp.Message))
	}

	return apiResp.Result, nil
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
		errMsg := "unknown error"
		if taskResult.Error != nil {
			errMsg = *taskResult.Error
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

	if taskResult.Error != nil {
		result.Error = *taskResult.Error
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
