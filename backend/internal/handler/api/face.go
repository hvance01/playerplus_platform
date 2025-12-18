package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"playplus_platform/internal/config"
	"playplus_platform/internal/service"
)

type DetectFacesRequest struct {
	ImageURL string `json:"image_url" binding:"required"`
}

type DetectedFaceResponse struct {
	Index     int       `json:"index"`
	FaceID    int       `json:"face_id"`
	Thumbnail string    `json:"thumbnail,omitempty"`
	BBox      []float64 `json:"bbox,omitempty"`
}

// DetectFacesResponseData - ALL responses include Status for polling logic
type DetectFacesResponseData struct {
	TaskID     string                 `json:"task_id,omitempty"`
	Status     string                 `json:"status"` // queuing, processing, completed, failed
	Faces      []DetectedFaceResponse `json:"faces"`
	DetectID   string                 `json:"detect_id,omitempty"`
	FrameImage string                 `json:"frame_image,omitempty"`
}

type DetectFacesResponse struct {
	Code int                      `json:"code"`
	Data *DetectFacesResponseData `json:"data,omitempty"`
	Msg  string                   `json:"msg,omitempty"`
}

// DetectFaces starts face detection from an image URL
// Production: returns task_id with status="queuing", client polls GetFaceDetectStatus
// Mock mode: returns completed response immediately
func DetectFaces(c *gin.Context) {
	var req DetectFacesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, DetectFacesResponse{
			Code: 400,
			Msg:  "Invalid request: " + err.Error(),
		})
		return
	}

	cfg := config.Get()

	if cfg.IsVModelConfigured() {
		detectWithVModel(c, req.ImageURL)
		return
	}

	// Mock mode: return completed immediately (backward compatible)
	c.JSON(http.StatusOK, DetectFacesResponse{
		Code: 0,
		Data: &DetectFacesResponseData{
			Status:     "completed",
			Faces:      []DetectedFaceResponse{{Index: 0, FaceID: 0, BBox: []float64{100, 100, 150, 150}}},
			DetectID:   "mock_detect_id",
			FrameImage: req.ImageURL,
		},
		Msg: "Mock response - VModel API not configured",
	})
}

// detectWithVModel starts async face detection using VModel API
func detectWithVModel(c *gin.Context, imageURL string) {
	// Convert CDN URL to direct storage URL for VModel API
	storage := service.GetStorageService()
	directURL := storage.ConvertToDirectURL(imageURL)
	log.Printf("[DEBUG] detectWithVModel: CDN=%s, Direct=%s", imageURL, directURL)

	vmodel := service.GetVModelClient()
	result, err := vmodel.CreateDetectTask(c.Request.Context(), directURL)
	if err != nil {
		log.Printf("[ERROR] VModel create detect task failed: %v", err)
		c.JSON(http.StatusInternalServerError, DetectFacesResponse{
			Code: 500,
			Msg:  "Failed to start face detection: " + err.Error(),
		})
		return
	}

	log.Printf("[INFO] Face detection task created: %s", result.TaskID)
	c.JSON(http.StatusOK, DetectFacesResponse{
		Code: 0,
		Data: &DetectFacesResponseData{
			TaskID:     result.TaskID,
			Status:     result.Status, // "queuing"
			Faces:      []DetectedFaceResponse{},
			FrameImage: imageURL, // Frontend stores this for later use
		},
	})
}

// GetFaceDetectStatus returns the status of a face detection task
// Client should poll this endpoint until status is "completed" or "failed"
func GetFaceDetectStatus(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, DetectFacesResponse{
			Code: 400,
			Msg:  "Task ID is required",
		})
		return
	}

	cfg := config.Get()
	if !cfg.IsVModelConfigured() {
		// Mock: return completed immediately
		c.JSON(http.StatusOK, DetectFacesResponse{
			Code: 0,
			Data: &DetectFacesResponseData{
				TaskID:   taskID,
				Status:   "completed",
				Faces:    []DetectedFaceResponse{{Index: 0, FaceID: 0}},
				DetectID: "mock_detect_id",
			},
			Msg: "Mock response",
		})
		return
	}

	vmodel := service.GetVModelClient()
	result, err := vmodel.GetDetectTaskStatus(c.Request.Context(), taskID)

	// HTTP/decode errors (result is nil)
	if err != nil && result == nil {
		log.Printf("[ERROR] VModel get detect status failed: %v", err)
		c.JSON(http.StatusInternalServerError, DetectFacesResponse{
			Code: 500,
			Msg:  "Failed to get detection status: " + err.Error(),
		})
		return
	}

	// Queuing/processing - client should continue polling
	if result.Status == "queuing" || result.Status == "processing" {
		c.JSON(http.StatusOK, DetectFacesResponse{
			Code: 0,
			Data: &DetectFacesResponseData{
				TaskID: taskID,
				Status: result.Status,
				Faces:  []DetectedFaceResponse{},
			},
		})
		return
	}

	// Failed - terminal state
	if result.Status == "failed" {
		log.Printf("[WARN] Face detection failed for task %s: %s", taskID, result.Error)
		c.JSON(http.StatusOK, DetectFacesResponse{
			Code: 0,
			Data: &DetectFacesResponseData{
				TaskID: taskID,
				Status: "failed",
				Faces:  []DetectedFaceResponse{},
			},
			Msg: result.Error,
		})
		return
	}

	// Completed - build faces response
	faces := make([]DetectedFaceResponse, len(result.Faces))
	for i, f := range result.Faces {
		faces[i] = DetectedFaceResponse{
			Index:     i,
			FaceID:    f.ID,
			Thumbnail: f.Link,
		}
	}

	log.Printf("[INFO] Face detection completed for task %s: %d faces found", taskID, len(faces))
	c.JSON(http.StatusOK, DetectFacesResponse{
		Code: 0,
		Data: &DetectFacesResponseData{
			TaskID:   taskID,
			Status:   "completed",
			Faces:    faces,
			DetectID: result.DetectID,
		},
	})
}

// DetectFacesFromUpload handles face detection from uploaded image (form-data)
func DetectFacesFromUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, DetectFacesResponse{
			Code: 400,
			Msg:  "No file uploaded",
		})
		return
	}

	// Upload the frame first
	storage := service.GetStorageService()
	key := storage.GenerateKey("frames", file.Filename)
	url, err := storage.Upload(c.Request.Context(), key, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, DetectFacesResponse{
			Code: 500,
			Msg:  "Failed to upload frame: " + err.Error(),
		})
		return
	}

	cfg := config.Get()

	if cfg.IsVModelConfigured() {
		detectWithVModel(c, url)
		return
	}

	// Mock mode: return completed immediately
	c.JSON(http.StatusOK, DetectFacesResponse{
		Code: 0,
		Data: &DetectFacesResponseData{
			Status:     "completed",
			Faces:      []DetectedFaceResponse{{Index: 0, FaceID: 0}},
			DetectID:   "mock_detect_id",
			FrameImage: url,
		},
		Msg: "Mock response - VModel API not configured",
	})
}
