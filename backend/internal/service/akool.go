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

// AkoolClient handles all Akool API interactions
type AkoolClient struct {
	httpClient *http.Client
	cfg        *config.Config
	token      string
	tokenExp   time.Time
	tokenMu    sync.RWMutex
}

var (
	akoolClient *AkoolClient
	akoolOnce   sync.Once
)

// GetAkoolClient returns the singleton Akool client
func GetAkoolClient() *AkoolClient {
	akoolOnce.Do(func() {
		akoolClient = &AkoolClient{
			httpClient: &http.Client{Timeout: 60 * time.Second},
			cfg:        config.Get(),
		}
	})
	return akoolClient
}

// --- Token Management ---

type tokenResponse struct {
	Code int `json:"code"`
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
	Msg string `json:"msg"`
}

// GetToken retrieves or refreshes the API token
func (c *AkoolClient) GetToken(ctx context.Context) (string, error) {
	c.tokenMu.RLock()
	if c.token != "" && time.Now().Before(c.tokenExp) {
		token := c.token
		c.tokenMu.RUnlock()
		return token, nil
	}
	c.tokenMu.RUnlock()

	// Refresh token
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	// Double-check after acquiring write lock
	if c.token != "" && time.Now().Before(c.tokenExp) {
		return c.token, nil
	}

	url := fmt.Sprintf("%s/api/open/v3/getToken", c.cfg.AkoolBaseURL)
	body := map[string]string{
		"clientId":     c.cfg.AkoolClientID,
		"clientSecret": c.cfg.AkoolAPIKey,
	}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	var result tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}

	if result.Code != 1000 {
		return "", fmt.Errorf("token error: %s", result.Msg)
	}

	c.token = result.Data.Token
	c.tokenExp = time.Now().Add(24 * time.Hour) // Token valid for ~1 year, refresh daily

	return c.token, nil
}

// --- Face Detection ---

type DetectFaceRequest struct {
	ImageURL   string `json:"image_url,omitempty"`
	ImageBase64 string `json:"img,omitempty"`
	SingleFace bool   `json:"single_face"`
}

type DetectedFace struct {
	Index        int       `json:"index"`
	BBox         []float64 `json:"bbox"`           // [x, y, width, height]
	LandmarksStr string    `json:"landmarks_str"`  // For Akool swap API
	Thumbnail    string    `json:"thumbnail"`      // Base64 cropped face image
}

type DetectFaceResponse struct {
	Faces      []DetectedFace `json:"faces"`
	FrameImage string         `json:"frame_image"` // URL of the frame used for detection
}

type akoolDetectResponse struct {
	ErrorCode int    `json:"error_code"`
	Msg       string `json:"msg"`
	Data      struct {
		Landmarks    [][][2]float64 `json:"landmarks"`
		LandmarksStr []string       `json:"landmarks_str"`
	} `json:"data"`
}

// DetectFaces detects faces in an image
func (c *AkoolClient) DetectFaces(ctx context.Context, imageURL string) (*DetectFaceResponse, error) {
	token, err := c.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}

	url := fmt.Sprintf("%s/detect", c.cfg.AkoolDetectURL)
	body := DetectFaceRequest{
		ImageURL:   imageURL,
		SingleFace: false, // Detect all faces
	}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create detect request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("detect request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result akoolDetectResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decode detect response: %w (body: %s)", err, string(respBody))
	}

	if result.ErrorCode != 0 {
		return nil, fmt.Errorf("detect error: %s", result.Msg)
	}

	// Build response with face info
	faces := make([]DetectedFace, 0, len(result.Data.LandmarksStr))
	for i, landmarksStr := range result.Data.LandmarksStr {
		face := DetectedFace{
			Index:        i,
			LandmarksStr: landmarksStr,
		}

		// Calculate bounding box from landmarks if available
		if i < len(result.Data.Landmarks) && len(result.Data.Landmarks[i]) > 0 {
			landmarks := result.Data.Landmarks[i]
			minX, minY := landmarks[0][0], landmarks[0][1]
			maxX, maxY := landmarks[0][0], landmarks[0][1]
			for _, pt := range landmarks {
				if pt[0] < minX {
					minX = pt[0]
				}
				if pt[0] > maxX {
					maxX = pt[0]
				}
				if pt[1] < minY {
					minY = pt[1]
				}
				if pt[1] > maxY {
					maxY = pt[1]
				}
			}
			// Add padding
			padding := (maxX - minX) * 0.2
			face.BBox = []float64{
				minX - padding,
				minY - padding,
				(maxX - minX) + 2*padding,
				(maxY - minY) + 2*padding,
			}
		}

		faces = append(faces, face)
	}

	return &DetectFaceResponse{
		Faces:      faces,
		FrameImage: imageURL,
	}, nil
}

// --- Face Swap ---

type FaceSwapPair struct {
	SourceImageURL string `json:"source_image_url"` // New face to use
	LandmarksStr   string `json:"landmarks_str"`    // Target face to replace
}

type CreateSwapTaskRequest struct {
	TargetVideoURL string         `json:"target_video_url"`
	FaceSwaps      []FaceSwapPair `json:"face_swaps"`
	FaceEnhance    bool           `json:"face_enhance"`
	WebhookURL     string         `json:"webhook_url,omitempty"`
}

type SwapTaskResult struct {
	JobID     string `json:"job_id"`
	Status    int    `json:"status"` // 1=queuing, 2=processing, 3=completed, 4=failed
	ResultURL string `json:"result_url,omitempty"`
	Error     string `json:"error,omitempty"`
}

type akoolSwapResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		ID string `json:"_id"`
	} `json:"data"`
}

// CreateSwapTask creates a video face swap task using V3 API
func (c *AkoolClient) CreateSwapTask(ctx context.Context, req *CreateSwapTaskRequest) (*SwapTaskResult, error) {
	token, err := c.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}

	// Build sourceImage and targetImage arrays
	sourceImages := make([]map[string]string, 0, len(req.FaceSwaps))
	targetImages := make([]map[string]string, 0, len(req.FaceSwaps))

	// We need a frame image URL for targetImage - use first face swap's context
	// In real implementation, this should be the frame used for detection
	frameURL := req.TargetVideoURL // Will be replaced with actual frame

	for _, swap := range req.FaceSwaps {
		sourceImages = append(sourceImages, map[string]string{
			"path": swap.SourceImageURL,
		})
		targetImages = append(targetImages, map[string]string{
			"path": frameURL,
			"opts": swap.LandmarksStr,
		})
	}

	url := fmt.Sprintf("%s/api/open/v3/faceswap/highquality/specifyimage/createbyvideo", c.cfg.AkoolBaseURL)
	body := map[string]interface{}{
		"sourceImage":  sourceImages,
		"targetImage":  targetImages,
		"modifyVideo":  req.TargetVideoURL,
		"face_enhance": req.FaceEnhance,
	}
	if req.WebhookURL != "" {
		body["webhookUrl"] = req.WebhookURL
	}

	jsonBody, _ := json.Marshal(body)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create swap request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("swap request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result akoolSwapResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decode swap response: %w (body: %s)", err, string(respBody))
	}

	if result.Code != 1000 {
		return nil, fmt.Errorf("swap error: %s", result.Msg)
	}

	return &SwapTaskResult{
		JobID:  result.Data.ID,
		Status: 1, // Queuing
	}, nil
}

// --- Task Status ---

type akoolTaskResultResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Result []struct {
			ID           string `json:"_id"`
			FaceswapType int    `json:"faceswap_type"`
			Status       int    `json:"status"` // 1=queuing, 2=processing, 3=success, 4=failed
			URL          string `json:"url"`
			Msg          string `json:"msg"`
		} `json:"result"`
	} `json:"data"`
}

// GetTaskStatus retrieves the status of a face swap task
func (c *AkoolClient) GetTaskStatus(ctx context.Context, jobID string) (*SwapTaskResult, error) {
	token, err := c.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}

	url := fmt.Sprintf("%s/api/open/v3/faceswap/result/listbyids?_ids=%s", c.cfg.AkoolBaseURL, jobID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create status request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("status request failed: %w", err)
	}
	defer resp.Body.Close()

	var result akoolTaskResultResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode status response: %w", err)
	}

	if result.Code != 1000 {
		return nil, fmt.Errorf("status error: %s", result.Msg)
	}

	if len(result.Data.Result) == 0 {
		return nil, fmt.Errorf("task not found: %s", jobID)
	}

	task := result.Data.Result[0]
	return &SwapTaskResult{
		JobID:     task.ID,
		Status:    task.Status,
		ResultURL: task.URL,
		Error:     task.Msg,
	}, nil
}

// StatusToString converts Akool status code to string
func StatusToString(status int) string {
	switch status {
	case 1:
		return "queuing"
	case 2:
		return "processing"
	case 3:
		return "completed"
	case 4:
		return "failed"
	default:
		return "unknown"
	}
}
