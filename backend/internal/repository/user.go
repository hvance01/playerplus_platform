package repository

import (
	"context"
	"database/sql"
	"time"
)

type User struct {
	ID          int64
	Email       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastLoginAt sql.NullTime
}

type VerificationCode struct {
	ID        int64
	Email     string
	Code      string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}

type Session struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// CreateOrGetUser creates a user if not exists, returns user ID
func CreateOrGetUser(ctx context.Context, email string) (int64, error) {
	if !IsDBAvailable() {
		return 0, nil
	}

	var userID int64
	err := db.QueryRowContext(ctx, `
		INSERT INTO users (email) VALUES ($1)
		ON CONFLICT (email) DO UPDATE SET updated_at = NOW()
		RETURNING id
	`, email).Scan(&userID)

	return userID, err
}

// SaveVerificationCode saves a verification code
func SaveVerificationCode(ctx context.Context, email, code string, expiresAt time.Time) error {
	if !IsDBAvailable() {
		return nil
	}

	// Invalidate previous codes for this email
	_, err := db.ExecContext(ctx, `
		UPDATE verification_codes SET used = TRUE WHERE email = $1 AND used = FALSE
	`, email)
	if err != nil {
		return err
	}

	// Insert new code
	_, err = db.ExecContext(ctx, `
		INSERT INTO verification_codes (email, code, expires_at) VALUES ($1, $2, $3)
	`, email, code, expiresAt)

	return err
}

// VerifyCode checks and marks code as used
func VerifyCodeDB(ctx context.Context, email, code string) (bool, error) {
	if !IsDBAvailable() {
		return false, nil
	}

	result, err := db.ExecContext(ctx, `
		UPDATE verification_codes
		SET used = TRUE
		WHERE email = $1 AND code = $2 AND expires_at > NOW() AND used = FALSE
	`, email, code)
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	return rows > 0, err
}

// CreateSession creates a new session
func CreateSession(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	if !IsDBAvailable() {
		return nil
	}

	_, err := db.ExecContext(ctx, `
		INSERT INTO sessions (user_id, token, expires_at) VALUES ($1, $2, $3)
	`, userID, token, expiresAt)

	return err
}

// GetSessionByToken retrieves session by token
func GetSessionByToken(ctx context.Context, token string) (*Session, error) {
	if !IsDBAvailable() {
		return nil, nil
	}

	var s Session
	err := db.QueryRowContext(ctx, `
		SELECT id, user_id, token, expires_at, created_at
		FROM sessions
		WHERE token = $1 AND expires_at > NOW()
	`, token).Scan(&s.ID, &s.UserID, &s.Token, &s.ExpiresAt, &s.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

// UpdateUserLastLogin updates user's last login time
func UpdateUserLastLogin(ctx context.Context, userID int64) error {
	if !IsDBAvailable() {
		return nil
	}

	_, err := db.ExecContext(ctx, `
		UPDATE users SET last_login_at = NOW() WHERE id = $1
	`, userID)

	return err
}

// CleanupExpiredCodes removes expired verification codes
func CleanupExpiredCodes(ctx context.Context) error {
	if !IsDBAvailable() {
		return nil
	}

	_, err := db.ExecContext(ctx, `
		DELETE FROM verification_codes WHERE expires_at < NOW() - INTERVAL '1 day'
	`)
	return err
}

// CleanupExpiredSessions removes expired sessions
func CleanupExpiredSessions(ctx context.Context) error {
	if !IsDBAvailable() {
		return nil
	}

	_, err := db.ExecContext(ctx, `
		DELETE FROM sessions WHERE expires_at < NOW()
	`)
	return err
}
