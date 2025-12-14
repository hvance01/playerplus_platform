package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"playplus_platform/internal/config"
	"playplus_platform/internal/service"
)

// --- Request/Response Types ---

// FaceSwapPairRequest supports both VModel (face_id) and Akool (landmarks_str)
type FaceSwapPairRequest struct {
	SourceImageURL string `json:"source_image_url" binding:"required"` // New face image URL
	FaceID         int    `json:"face_id"`                             // VModel: target face ID
	LandmarksStr   string `json:"landmarks_str"`                       // Akool: target face landmarks (legacy)
}

type CreateFaceSwapRequest struct {
	TargetVideoURL string                `json:"target_video_url" binding:"required"`
	DetectID       string                `json:"detect_id"`                            // VModel: detection ID
	FrameImageURL  string                `json:"frame_image_url"`                      // Akool: frame used for detection
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
		TaskID    string `json:"task_id"`
		Status    string `json:"status"` // queuing, processing, completed, failed
		ResultURL string `json:"result_url,omitempty"`
		Error     string `json:"error,omitempty"`
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

	// Try VModel first (preferred)
	if cfg.IsVModelConfigured() {
		createSwapWithVModel(c, &req)
		return
	}

	// Fallback to Akool
	if cfg.IsAkoolConfigured() {
		createSwapWithAkool(c, &req)
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
		Msg: "Mock response - No face swap API configured",
	})
}

// createSwapWithVModel creates swap task using VModel API
func createSwapWithVModel(c *gin.Context, req *CreateFaceSwapRequest) {
	if req.DetectID == "" {
		c.JSON(http.StatusBadRequest, CreateFaceSwapResponse{
			Code: 400,
			Msg:  "detect_id is required for VModel API",
		})
		return
	}

	// Build face swap pairs
	faceSwaps := make([]service.VModelFaceSwapPair, len(req.FaceSwaps))
	for i, swap := range req.FaceSwaps {
		faceSwaps[i] = service.VModelFaceSwapPair{
			FaceID: swap.FaceID,
			Target: swap.SourceImageURL,
		}
	}

	vmodel := service.GetVModelClient()
	result, err := vmodel.CreateSwapTask(c.Request.Context(), req.DetectID, faceSwaps)
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

// createSwapWithAkool creates swap task using Akool API (legacy)
func createSwapWithAkool(c *gin.Context, req *CreateFaceSwapRequest) {
	// Build Akool request
	faceSwaps := make([]service.FaceSwapPair, len(req.FaceSwaps))
	for i, swap := range req.FaceSwaps {
		faceSwaps[i] = service.FaceSwapPair{
			SourceImageURL: swap.SourceImageURL,
			LandmarksStr:   swap.LandmarksStr,
		}
	}

	akool := service.GetAkoolClient()
	result, err := akool.CreateSwapTask(c.Request.Context(), &service.CreateSwapTaskRequest{
		TargetVideoURL: req.TargetVideoURL,
		FaceSwaps:      faceSwaps,
		FaceEnhance:    req.FaceEnhance,
	})
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
			TaskID: result.JobID,
			Status: service.StatusToString(result.Status),
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

	// Try VModel first (preferred)
	if cfg.IsVModelConfigured() {
		getStatusFromVModel(c, taskID)
		return
	}

	// Fallback to Akool
	if cfg.IsAkoolConfigured() {
		getStatusFromAkool(c, taskID)
		return
	}

	// Return mock response
	c.JSON(http.StatusOK, GetTaskStatusResponse{
		Code: 0,
		Data: &struct {
			TaskID    string `json:"task_id"`
			Status    string `json:"status"`
			ResultURL string `json:"result_url,omitempty"`
			Error     string `json:"error,omitempty"`
		}{
			TaskID:    taskID,
			Status:    "completed",
			ResultURL: "https://mock.example.com/result.mp4",
		},
		Msg: "Mock response - No face swap API configured",
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

	c.JSON(http.StatusOK, GetTaskStatusResponse{
		Code: 0,
		Data: &struct {
			TaskID    string `json:"task_id"`
			Status    string `json:"status"`
			ResultURL string `json:"result_url,omitempty"`
			Error     string `json:"error,omitempty"`
		}{
			TaskID:    result.TaskID,
			Status:    result.Status,
			ResultURL: result.ResultURL,
			Error:     result.Error,
		},
	})
}

// getStatusFromAkool gets task status from Akool API
func getStatusFromAkool(c *gin.Context, taskID string) {
	akool := service.GetAkoolClient()
	result, err := akool.GetTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GetTaskStatusResponse{
			Code: 500,
			Msg:  "Failed to get task status: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GetTaskStatusResponse{
		Code: 0,
		Data: &struct {
			TaskID    string `json:"task_id"`
			Status    string `json:"status"`
			ResultURL string `json:"result_url,omitempty"`
			Error     string `json:"error,omitempty"`
		}{
			TaskID:    result.JobID,
			Status:    service.StatusToString(result.Status),
			ResultURL: result.ResultURL,
			Error:     result.Error,
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
	var oldReq struct {
		MediaID string   `json:"media_id" binding:"required"`
		FaceIDs []string `json:"face_ids" binding:"required,min=1"`
		Model   string   `json:"model" binding:"required"`
	}
	if err := c.ShouldBindJSON(&oldReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id": "legacy_mock_task",
		"status":  "processing",
		"message": "Please migrate to /api/v2/faceswap/create",
	})
}

// GetTaskStatus is kept for backward compatibility
func GetTaskStatus(c *gin.Context) {
	GetFaceSwapTaskStatus(c)
}
