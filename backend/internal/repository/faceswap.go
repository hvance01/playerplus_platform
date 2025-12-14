package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type MediaFile struct {
	ID           int64
	UserID       int64
	MediaID      string
	Filename     string
	FileType     string
	FileSize     int64
	StorageURL   sql.NullString
	ThumbnailURL sql.NullString
	CreatedAt    time.Time
}

type SwapTask struct {
	ID           int64
	UserID       int64
	TaskID       string
	MediaID      string
	FaceIDs      []string
	Model        string
	Status       string
	ResultURL    sql.NullString
	ErrorMessage sql.NullString
	CreditsUsed  float64
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CompletedAt  sql.NullTime
}

// SaveMediaFile saves a media file record
func SaveMediaFile(ctx context.Context, userID int64, mediaID, filename, fileType string, fileSize int64) error {
	if !IsDBAvailable() {
		return nil
	}

	_, err := db.ExecContext(ctx, `
		INSERT INTO media_files (user_id, media_id, filename, file_type, file_size)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, mediaID, filename, fileType, fileSize)

	return err
}

// GetMediaFilesByUser retrieves media files for a user
func GetMediaFilesByUser(ctx context.Context, userID int64, limit int) ([]MediaFile, error) {
	if !IsDBAvailable() {
		return nil, nil
	}

	rows, err := db.QueryContext(ctx, `
		SELECT id, user_id, media_id, filename, file_type, file_size, storage_url, thumbnail_url, created_at
		FROM media_files
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []MediaFile
	for rows.Next() {
		var f MediaFile
		err := rows.Scan(&f.ID, &f.UserID, &f.MediaID, &f.Filename, &f.FileType, &f.FileSize, &f.StorageURL, &f.ThumbnailURL, &f.CreatedAt)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	return files, rows.Err()
}

// CreateSwapTask creates a new swap task
func CreateSwapTask(ctx context.Context, userID int64, taskID, mediaID string, faceIDs []string, model string) error {
	if !IsDBAvailable() {
		return nil
	}

	_, err := db.ExecContext(ctx, `
		INSERT INTO swap_tasks (user_id, task_id, media_id, face_ids, model, status)
		VALUES ($1, $2, $3, $4, $5, 'pending')
	`, userID, taskID, mediaID, pq.Array(faceIDs), model)

	return err
}

// GetSwapTask retrieves a swap task by task ID
func GetSwapTask(ctx context.Context, taskID string) (*SwapTask, error) {
	if !IsDBAvailable() {
		return nil, nil
	}

	var t SwapTask
	err := db.QueryRowContext(ctx, `
		SELECT id, user_id, task_id, media_id, face_ids, model, status,
		       result_url, error_message, credits_used, created_at, updated_at, completed_at
		FROM swap_tasks
		WHERE task_id = $1
	`, taskID).Scan(
		&t.ID, &t.UserID, &t.TaskID, &t.MediaID, pq.Array(&t.FaceIDs), &t.Model, &t.Status,
		&t.ResultURL, &t.ErrorMessage, &t.CreditsUsed, &t.CreatedAt, &t.UpdatedAt, &t.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

// UpdateSwapTaskStatus updates task status
func UpdateSwapTaskStatus(ctx context.Context, taskID, status string, resultURL, errorMsg *string) error {
	if !IsDBAvailable() {
		return nil
	}

	query := `UPDATE swap_tasks SET status = $2`
	args := []interface{}{taskID, status}

	if resultURL != nil {
		query += `, result_url = $3`
		args = append(args, *resultURL)
	}
	if errorMsg != nil {
		if resultURL != nil {
			query += `, error_message = $4`
		} else {
			query += `, error_message = $3`
		}
		args = append(args, *errorMsg)
	}
	if status == "completed" || status == "failed" {
		query += `, completed_at = NOW()`
	}

	query += ` WHERE task_id = $1`

	_, err := db.ExecContext(ctx, query, args...)
	return err
}

// GetSwapTasksByUser retrieves tasks for a user
func GetSwapTasksByUser(ctx context.Context, userID int64, limit int) ([]SwapTask, error) {
	if !IsDBAvailable() {
		return nil, nil
	}

	rows, err := db.QueryContext(ctx, `
		SELECT id, user_id, task_id, media_id, face_ids, model, status,
		       result_url, error_message, credits_used, created_at, updated_at, completed_at
		FROM swap_tasks
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []SwapTask
	for rows.Next() {
		var t SwapTask
		err := rows.Scan(
			&t.ID, &t.UserID, &t.TaskID, &t.MediaID, pq.Array(&t.FaceIDs), &t.Model, &t.Status,
			&t.ResultURL, &t.ErrorMessage, &t.CreditsUsed, &t.CreatedAt, &t.UpdatedAt, &t.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, rows.Err()
}
