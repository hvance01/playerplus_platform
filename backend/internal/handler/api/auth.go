package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"playplus_platform/internal/service"
)

type SendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

// SendVerificationCode sends a verification code to the email
func SendVerificationCode(c *gin.Context) {
	var req SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validate email domain
	if !strings.HasSuffix(strings.ToLower(req.Email), "@playerplus.cn") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only @playerplus.cn emails are allowed"})
		return
	}

	if err := service.SendVerificationCode(req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Verification code sent"})
}

// VerifyCode verifies the code and returns a session token
func VerifyCode(c *gin.Context) {
	var req VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	token, err := service.VerifyCode(req.Email, req.Code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
