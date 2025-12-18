package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/dm"
	"playplus_platform/internal/config"
	"playplus_platform/internal/repository"
)

var (
	// In-memory store for verification codes (fallback when no DB)
	codeStore = make(map[string]codeEntry)
	codeMu    sync.RWMutex

	dmClient *dm.Client
)

func init() {
	cfg := config.Get()
	if cfg.AliyunAccessKeyID != "" && cfg.AliyunAccessKeySecret != "" {
		client, err := dm.NewClientWithAccessKey(cfg.AliyunEmailRegion, cfg.AliyunAccessKeyID, cfg.AliyunAccessKeySecret)
		if err != nil {
			fmt.Printf("[ERROR] Failed to create Aliyun DM client: %v\n", err)
		} else {
			dmClient = client
			fmt.Printf("[INFO] Aliyun DirectMail client initialized\n")
		}
	}
}

type codeEntry struct {
	Code      string
	ExpiresAt time.Time
}

// SendVerificationCode generates and sends a verification code
func SendVerificationCode(email string) error {
	code := generateCode()
	expiresAt := time.Now().Add(10 * time.Minute)

	// Save to database if available
	ctx := context.Background()
	if repository.IsDBAvailable() {
		if err := repository.SaveVerificationCode(ctx, email, code, expiresAt); err != nil {
			fmt.Printf("[ERROR] Failed to save code to DB: %v\n", err)
		}
	} else {
		// Fallback to in-memory
		codeMu.Lock()
		codeStore[email] = codeEntry{
			Code:      code,
			ExpiresAt: expiresAt,
		}
		codeMu.Unlock()
	}

	// Send email via Aliyun DirectMail
	if dmClient != nil {
		cfg := config.Get()
		request := dm.CreateSingleSendMailRequest()
		request.Scheme = "https"
		request.AccountName = cfg.AliyunEmailFrom
		request.FromAlias = "PlayerPlus"
		request.AddressType = "1"
		request.ReplyToAddress = "false"
		request.ToAddress = email
		request.Subject = "PlayerPlus 登录验证码"
		request.HtmlBody = fmt.Sprintf(`
			<div style="font-family: sans-serif; max-width: 600px; margin: 0 auto;">
				<h2 style="color: #1890ff;">PlayerPlus Platform</h2>
				<p>您好，</p>
				<p>您的登录验证码是：</p>
				<div style="background: #f5f5f5; padding: 20px; text-align: center; margin: 20px 0;">
					<span style="font-size: 32px; font-weight: bold; letter-spacing: 8px; color: #333;">%s</span>
				</div>
				<p>验证码有效期为 10 分钟，请勿泄露给他人。</p>
				<p style="color: #999; font-size: 12px;">如果您没有请求此验证码，请忽略此邮件。</p>
			</div>
		`, code)

		_, err := dmClient.SingleSendMail(request)
		if err != nil {
			fmt.Printf("[ERROR] Failed to send email via Aliyun: %v\n", err)
			fmt.Printf("[DEV] Verification code for %s: %s\n", email, code)
			return nil
		}
		fmt.Printf("[INFO] Verification code sent to %s via Aliyun DirectMail\n", email)
	} else {
		fmt.Printf("[DEV] Verification code for %s: %s (Aliyun DM not configured)\n", email, code)
	}

	return nil
}

// VerifyCode checks the code and returns a session token
func VerifyCode(email, code string) (string, error) {
	ctx := context.Background()

	// Try database first
	if repository.IsDBAvailable() {
		valid, err := repository.VerifyCodeDB(ctx, email, code)
		if err != nil {
			return "", err
		}
		if !valid {
			return "", errors.New("invalid or expired code")
		}

		// Create or get user
		userID, err := repository.CreateOrGetUser(ctx, email)
		if err != nil {
			return "", err
		}

		// Update last login
		repository.UpdateUserLastLogin(ctx, userID)

		// Generate and save session token
		token := generateToken()
		expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
		if err := repository.CreateSession(ctx, userID, token, expiresAt); err != nil {
			return "", err
		}

		return token, nil
	}

	// Fallback to in-memory
	codeMu.RLock()
	entry, exists := codeStore[email]
	codeMu.RUnlock()

	if !exists {
		return "", errors.New("no code found")
	}

	if time.Now().After(entry.ExpiresAt) {
		return "", errors.New("code expired")
	}

	if entry.Code != code {
		return "", errors.New("invalid code")
	}

	codeMu.Lock()
	delete(codeStore, email)
	codeMu.Unlock()

	token := generateToken()
	return token, nil
}

func generateCode() string {
	b := make([]byte, 3)
	rand.Read(b)
	return fmt.Sprintf("%06d", int(b[0])*10000+int(b[1])*100+int(b[2])%100)[:6]
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
