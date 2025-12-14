package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"playplus_platform/internal/middleware"
	"playplus_platform/internal/service"
)

// UploadMediaFile handles video/image file upload
func UploadMediaFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	if !isValidMediaType(contentType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only images and videos are allowed"})
		return
	}

	// Generate storage key
	storage := service.GetStorageService()
	prefix := "videos"
	if isImageType(contentType) {
		prefix = "images"
	}
	key := storage.GenerateKey(prefix, file.Filename)

	// Upload to storage
	url, err := storage.Upload(c.Request.Context(), key, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file: " + err.Error()})
		return
	}

	// Get user ID if authenticated
	userID := middleware.GetUserID(c)

	c.JSON(http.StatusOK, gin.H{
		"url":          url,
		"key":          key,
		"filename":     file.Filename,
		"content_type": contentType,
		"size":         file.Size,
		"user_id":      userID,
	})
}

// UploadFaceImage handles face image upload (for replacement)
func UploadFaceImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Validate file type - only images for faces
	contentType := file.Header.Get("Content-Type")
	if !isImageType(contentType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only images are allowed for face upload"})
		return
	}

	// Generate storage key
	storage := service.GetStorageService()
	key := storage.GenerateKey("faces", file.Filename)

	// Upload to storage
	url, err := storage.Upload(c.Request.Context(), key, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url":          url,
		"key":          key,
		"filename":     file.Filename,
		"content_type": contentType,
	})
}

// UploadFrame handles video frame image upload (for face detection)
func UploadFrame(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Generate storage key
	storage := service.GetStorageService()
	key := storage.GenerateKey("frames", file.Filename)

	// Upload to storage
	url, err := storage.Upload(c.Request.Context(), key, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url": url,
		"key": key,
	})
}

func isValidMediaType(contentType string) bool {
	return isImageType(contentType) || isVideoType(contentType)
}

func isImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
	}
	for _, t := range validTypes {
		if contentType == t {
			return true
		}
	}
	return false
}

func isVideoType(contentType string) bool {
	validTypes := []string{
		"video/mp4",
		"video/webm",
		"video/quicktime",
		"video/x-msvideo",
	}
	for _, t := range validTypes {
		if contentType == t {
			return true
		}
	}
	return false
}
