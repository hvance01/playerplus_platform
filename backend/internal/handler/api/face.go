package api

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"playplus_platform/internal/config"
	"playplus_platform/internal/service"
)

type DetectFacesRequest struct {
	ImageURL string `json:"image_url" binding:"required"`
}

// DetectedFaceResponse represents a detected face in API response
type DetectedFaceResponse struct {
	Index     int       `json:"index"`
	FaceID    int       `json:"face_id"`
	Thumbnail string    `json:"thumbnail,omitempty"`
	BBox      []float64 `json:"bbox,omitempty"`
}

type DetectFacesResponseData struct {
	Faces      []DetectedFaceResponse `json:"faces"`
	DetectID   string                 `json:"detect_id,omitempty"`
	FrameImage string                 `json:"frame_image"`
}

type DetectFacesResponse struct {
	Code int                      `json:"code"`
	Data *DetectFacesResponseData `json:"data,omitempty"`
	Msg  string                   `json:"msg,omitempty"`
}

// DetectFaces handles face detection from an image URL
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

	// Return mock data for development
	c.JSON(http.StatusOK, DetectFacesResponse{
		Code: 0,
		Data: &DetectFacesResponseData{
			Faces: []DetectedFaceResponse{
				{
					Index:  0,
					FaceID: 0,
					BBox:   []float64{100, 100, 150, 150},
				},
			},
			DetectID:   "mock_detect_id",
			FrameImage: req.ImageURL,
		},
		Msg: "Mock response - VModel API not configured",
	})
}

// detectWithVModel uses VModel API for face detection
func detectWithVModel(c *gin.Context, imageURL string) {
	// Convert CDN URL to direct storage URL for VModel API
	// VModel is hosted outside China and doesn't need CDN acceleration
	storage := service.GetStorageService()
	directURL := storage.ConvertToDirectURL(imageURL)
	log.Printf("[DEBUG] detectWithVModel: CDN=%s, Direct=%s", imageURL, directURL)

	vmodel := service.GetVModelClient()
	result, err := vmodel.DetectFaces(c.Request.Context(), directURL)
	if err != nil {
		log.Printf("[ERROR] VModel face detection failed: %v", err)

		// Check if it's a "no face detected" error - return 200 with empty faces instead of 500
		errStr := err.Error()
		if strings.Contains(errStr, "未检测到人脸") || strings.Contains(errStr, "no face") || strings.Contains(errStr, "Detect.Failed") {
			c.JSON(http.StatusOK, DetectFacesResponse{
				Code: 0,
				Data: &DetectFacesResponseData{
					Faces:      []DetectedFaceResponse{},
					FrameImage: imageURL,
				},
				Msg: "未检测到人脸，请调整视频位置重试",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, DetectFacesResponse{
			Code: 500,
			Msg:  "Face detection failed: " + err.Error(),
		})
		return
	}

	// Convert to response format
	faces := make([]DetectedFaceResponse, len(result.Faces))
	for i, f := range result.Faces {
		faces[i] = DetectedFaceResponse{
			Index:     i,
			FaceID:    f.ID,
			Thumbnail: f.Link,
		}
	}

	c.JSON(http.StatusOK, DetectFacesResponse{
		Code: 0,
		Data: &DetectFacesResponseData{
			Faces:      faces,
			DetectID:   result.DetectID,
			FrameImage: imageURL,
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

	// Return mock data
	c.JSON(http.StatusOK, DetectFacesResponse{
		Code: 0,
		Data: &DetectFacesResponseData{
			Faces: []DetectedFaceResponse{
				{
					Index:  0,
					FaceID: 0,
					BBox:   []float64{100, 100, 150, 150},
				},
			},
			DetectID:   "mock_detect_id",
			FrameImage: url,
		},
		Msg: "Mock response - VModel API not configured",
	})
}
