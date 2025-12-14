package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"playplus_platform/internal/repository"
	"playplus_platform/internal/service"
)

type SendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Fixed credentials for development
const (
	fixedUsername = "test"
	fixedPassword = "test"
)

// Login handles username/password login
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Check fixed credentials
	if req.Username != fixedUsername || req.Password != fixedPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate a simple token
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	// Save session to database if available
	if repository.IsDBAvailable() {
		ctx := context.Background()
		// Create or get user (use username as email for test user)
		userID, err := repository.CreateOrGetUser(ctx, req.Username+"@test.local")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		// Create session with 7-day expiry
		expiresAt := time.Now().Add(7 * 24 * time.Hour)
		if err := repository.CreateSession(ctx, userID, token, expiresAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  req.Username,
	})
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
