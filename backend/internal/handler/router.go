package handler

import (
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"playplus_platform/internal/handler/api"
	"playplus_platform/internal/middleware"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Serve local uploads in development
	r.Static("/uploads", "./uploads")

	// API routes
	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Auth routes
		auth := apiGroup.Group("/auth")
		{
			auth.POST("/login", api.Login)
			auth.POST("/send-code", api.SendVerificationCode)
			auth.POST("/verify", api.VerifyCode)
		}

		// Legacy face swap routes (v1 - mock)
		faceswap := apiGroup.Group("/faceswap")
		faceswap.Use(middleware.AuthRequired())
		{
			faceswap.POST("/upload", api.UploadMedia)
			faceswap.POST("/swap", api.SwapFace)
			faceswap.GET("/tasks/:id", api.GetTaskStatus)
		}

		// New API v2 routes
		v2 := apiGroup.Group("/v2")
		v2.Use(middleware.AuthRequired())
		{
			// Media upload
			media := v2.Group("/media")
			{
				media.POST("/upload", api.UploadMediaFile)        // Upload video/image
				media.POST("/upload/face", api.UploadFaceImage)   // Upload face image
				media.POST("/upload/frame", api.UploadFrame)      // Upload video frame
			}

			// Face detection
			face := v2.Group("/face")
			{
				face.POST("/detect", api.DetectFaces)             // Detect faces from URL
				face.POST("/detect/upload", api.DetectFacesFromUpload) // Detect faces from uploaded file
			}

			// Face swap
			swap := v2.Group("/faceswap")
			{
				swap.POST("/create", api.CreateFaceSwapTask)      // Create face swap task
				swap.GET("/task/:id", api.GetFaceSwapTaskStatus)  // Get task status
			}
		}
	}

	// Serve frontend static files
	setupStaticFiles(r)

	return r
}

func setupStaticFiles(r *gin.Engine) {
	// In production, serve embedded frontend files
	// In development, this won't be used (frontend runs on vite dev server)
	staticFS, err := fs.Sub(frontendFS, "dist")
	if err != nil {
		// No embedded files (development mode)
		return
	}

	// Check if index.html exists
	indexFile, err := staticFS.Open("index.html")
	if err != nil {
		// No index.html found
		return
	}
	indexFile.Close()

	// Serve static files
	r.StaticFS("/assets", http.FS(staticFS))

	// SPA fallback - serve index.html for all non-API routes
	r.NoRoute(func(c *gin.Context) {
		// Read index.html content and serve it directly
		content, err := fs.ReadFile(staticFS, "index.html")
		if err != nil {
			c.String(http.StatusNotFound, "Not Found")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", content)
	})
}
