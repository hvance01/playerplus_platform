package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"playplus_platform/internal/config"
	"playplus_platform/internal/service"
)

// --- Request/Response Types ---

type FaceSwapPairRequest struct {
	SourceImageURL string `json:"source_image_url" binding:"required"` // New face image URL
	FaceID         int    `json:"face_id"`                             // VModel: target face ID
}

type CreateFaceSwapRequest struct {
	TargetVideoURL string                `json:"target_video_url" binding:"required"`
	DetectID       string                `json:"detect_id" binding:"required"`
	FaceSwaps      []FaceSwapPairRequest `json:"face_swaps" binding:"required,min=1"`
	FaceEnhance    bool                  `json:"face_enhance"`
}

type CreateFaceSwapResponse struct {
	Code int    `json:"code"`
	Data *struct {
		TaskID string `json:"task_id"`
		Status string `json:"status"`
	} `json:"data,omitempty"`
	Msg string `json:"msg,omitempty"`
}

type GetTaskStatusResponse struct {
	Code int    `json:"code"`
	Data *struct {
		TaskID         string `json:"task_id"`
		Status         string `json:"status"` // queuing, processing, completed, failed
		ResultURL      string `json:"result_url,omitempty"`
		Error          string `json:"error,omitempty"`
		TransferStatus string `json:"transfer_status,omitempty"` // pending, completed, failed
		OriginalURL    string `json:"original_url,omitempty"`
	} `json:"data,omitempty"`
	Msg string `json:"msg,omitempty"`
}

// --- Handlers ---

// CreateFaceSwapTask creates a new face swap task
func CreateFaceSwapTask(c *gin.Context) {
	var req CreateFaceSwapRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, CreateFaceSwapResponse{
			Code: 400,
			Msg:  "Invalid request: " + err.Error(),
		})
		return
	}

	cfg := config.Get()

	if cfg.IsVModelConfigured() {
		createSwapWithVModel(c, &req)
		return
	}

	// Return mock response for development
	c.JSON(http.StatusOK, CreateFaceSwapResponse{
		Code: 0,
		Data: &struct {
			TaskID string `json:"task_id"`
			Status string `json:"status"`
		}{
			TaskID: "mock_task_123",
			Status: "queuing",
		},
		Msg: "Mock response - VModel API not configured",
	})
}

// createSwapWithVModel creates swap task using VModel API
func createSwapWithVModel(c *gin.Context, req *CreateFaceSwapRequest) {
	// Build face swap pairs
	faceSwaps := make([]service.VModelFaceSwapPair, len(req.FaceSwaps))
	for i, swap := range req.FaceSwaps {
		faceSwaps[i] = service.VModelFaceSwapPair{
			FaceID: swap.FaceID,
			Target: swap.SourceImageURL,
		}
	}

	vmodel := service.GetVModelClient()
	result, err := vmodel.CreateSwapTask(c.Request.Context(), req.DetectID, faceSwaps, req.FaceEnhance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CreateFaceSwapResponse{
			Code: 500,
			Msg:  "Failed to create task: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, CreateFaceSwapResponse{
		Code: 0,
		Data: &struct {
			TaskID string `json:"task_id"`
			Status string `json:"status"`
		}{
			TaskID: result.TaskID,
			Status: result.Status,
		},
	})
}

// GetFaceSwapTaskStatus returns the status of a face swap task
func GetFaceSwapTaskStatus(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, GetTaskStatusResponse{
			Code: 400,
			Msg:  "Task ID is required",
		})
		return
	}

	cfg := config.Get()

	if cfg.IsVModelConfigured() {
		getStatusFromVModel(c, taskID)
		return
	}

	// Return mock response
	c.JSON(http.StatusOK, GetTaskStatusResponse{
		Code: 0,
		Data: &struct {
			TaskID         string `json:"task_id"`
			Status         string `json:"status"`
			ResultURL      string `json:"result_url,omitempty"`
			Error          string `json:"error,omitempty"`
			TransferStatus string `json:"transfer_status,omitempty"`
			OriginalURL    string `json:"original_url,omitempty"`
		}{
			TaskID:    taskID,
			Status:    "completed",
			ResultURL: "https://mock.example.com/result.mp4",
		},
		Msg: "Mock response - VModel API not configured",
	})
}

// getStatusFromVModel gets task status from VModel API
func getStatusFromVModel(c *gin.Context, taskID string) {
	vmodel := service.GetVModelClient()
	result, err := vmodel.GetTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GetTaskStatusResponse{
			Code: 500,
			Msg:  "Failed to get task status: " + err.Error(),
		})
		return
	}

	// Handle video transfer logic
	resultURL := result.ResultURL
	transferStatus := ""
	originalURL := ""
	finalStatus := result.Status

	if result.Status == "completed" && result.ResultURL != "" {
		originalURL = result.ResultURL
		storage := service.GetStorageService()

		// Check transfer status
		ts := storage.GetTransferStatus(taskID)

		if ts != nil {
			transferStatus = ts.Status
			if ts.Status == "completed" {
				// Transfer done, use MinIO URL
				resultURL = ts.MinioURL
				finalStatus = "completed"
			} else if ts.Status == "pending" {
				// Transfer in progress, return "transferring" status
				finalStatus = "transferring"
			} else if ts.Status == "failed" {
				// Transfer failed, still return completed but with original URL
				// User can try to access original URL or retry
				finalStatus = "completed"
				transferStatus = "failed"
			}
		} else {
			// Not started yet, start it and return "transferring"
			storage.TransferFromVModel(taskID, result.ResultURL)
			transferStatus = "pending"
			finalStatus = "transferring"
		}
	}

	c.JSON(http.StatusOK, GetTaskStatusResponse{
		Code: 0,
		Data: &struct {
			TaskID         string `json:"task_id"`
			Status         string `json:"status"`
			ResultURL      string `json:"result_url,omitempty"`
			Error          string `json:"error,omitempty"`
			TransferStatus string `json:"transfer_status,omitempty"`
			OriginalURL    string `json:"original_url,omitempty"`
		}{
			TaskID:         result.TaskID,
			Status:         finalStatus,
			ResultURL:      resultURL,
			Error:          result.Error,
			TransferStatus: transferStatus,
			OriginalURL:    originalURL,
		},
	})
}

// --- Legacy handlers (keep for backward compatibility) ---

// UploadMedia is kept for backward compatibility
func UploadMedia(c *gin.Context) {
	UploadMediaFile(c)
}

// SwapFace is kept for backward compatibility
func SwapFace(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"error":   "deprecated",
		"message": "Please migrate to /api/v2/faceswap/create",
	})
}

// GetTaskStatus is kept for backward compatibility
func GetTaskStatus(c *gin.Context) {
	GetFaceSwapTaskStatus(c)
}
