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

	// API routes
	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Auth routes
		auth := apiGroup.Group("/auth")
		{
			auth.POST("/send-code", api.SendVerificationCode)
			auth.POST("/verify", api.VerifyCode)
		}

		// Face swap routes (mock for now)
		faceswap := apiGroup.Group("/faceswap")
		faceswap.Use(middleware.AuthRequired())
		{
			faceswap.POST("/upload", api.UploadMedia)
			faceswap.POST("/swap", api.SwapFace)
			faceswap.GET("/tasks/:id", api.GetTaskStatus)
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

	// Serve static files
	r.StaticFS("/assets", http.FS(staticFS))

	// SPA fallback - serve index.html for all non-API routes
	r.NoRoute(func(c *gin.Context) {
		c.FileFromFS("index.html", http.FS(staticFS))
	})
}
