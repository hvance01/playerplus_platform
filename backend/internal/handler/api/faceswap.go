package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"playplus_platform/internal/middleware"
	"playplus_platform/internal/service"
)

type SwapFaceRequest struct {
	MediaID string   `json:"media_id" binding:"required"`
	FaceIDs []string `json:"face_ids" binding:"required,min=1"`
	Model   string   `json:"model" binding:"required"`
}

// UploadMedia handles media file upload
func UploadMedia(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	userID := middleware.GetUserID(c)
	mediaID, err := service.UploadMedia(userID, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload media"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"media_id": mediaID,
		"filename": file.Filename,
	})
}

// SwapFace initiates a face swap task (mock implementation)
func SwapFace(c *gin.Context) {
	var req SwapFaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID := middleware.GetUserID(c)
	taskID, err := service.CreateSwapTask(userID, req.MediaID, req.FaceIDs, req.Model)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create swap task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id": taskID,
		"status":  "processing",
	})
}

// GetTaskStatus returns the status of a face swap task
func GetTaskStatus(c *gin.Context) {
	taskID := c.Param("id")

	task, err := service.GetTaskStatus(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}
