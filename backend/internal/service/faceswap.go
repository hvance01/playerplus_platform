package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"mime/multipart"
	"sync"
	"time"

	"playplus_platform/internal/repository"
)

// Mock implementation - replace with DeepSwap API when available

var (
	taskStore = make(map[string]*SwapTask)
	taskMu    sync.RWMutex
)

type SwapTask struct {
	ID        string    `json:"id"`
	MediaID   string    `json:"media_id"`
	FaceIDs   []string  `json:"face_ids"`
	Model     string    `json:"model"`
	Status    string    `json:"status"` // pending, processing, completed, failed
	ResultURL string    `json:"result_url,omitempty"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UploadMedia handles file upload (mock)
func UploadMedia(userID int64, file *multipart.FileHeader) (string, error) {
	mediaID := generateMediaID()

	// Save to database if available
	ctx := context.Background()
	if repository.IsDBAvailable() {
		fileType := "image"
		if ct := file.Header.Get("Content-Type"); len(ct) > 5 && ct[:5] == "video" {
			fileType = "video"
		}
		if err := repository.SaveMediaFile(ctx, userID, mediaID, file.Filename, fileType, file.Size); err != nil {
			fmt.Printf("[ERROR] Failed to save media to DB: %v\n", err)
		}
	}

	// TODO: Upload to cloud storage (S3, R2, etc.)
	fmt.Printf("[MOCK] Uploaded media: %s -> %s\n", file.Filename, mediaID)
	return mediaID, nil
}

// CreateSwapTask creates a new face swap task (mock)
func CreateSwapTask(userID int64, mediaID string, faceIDs []string, model string) (string, error) {
	taskID := generateTaskID()

	task := &SwapTask{
		ID:        taskID,
		MediaID:   mediaID,
		FaceIDs:   faceIDs,
		Model:     model,
		Status:    "processing",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database if available
	ctx := context.Background()
	if repository.IsDBAvailable() {
		if err := repository.CreateSwapTask(ctx, userID, taskID, mediaID, faceIDs, model); err != nil {
			fmt.Printf("[ERROR] Failed to save task to DB: %v\n", err)
		}
	}

	// Also keep in memory for quick access
	taskMu.Lock()
	taskStore[taskID] = task
	taskMu.Unlock()

	// Simulate async processing
	go simulateProcessing(taskID)

	fmt.Printf("[MOCK] Created swap task: %s (media: %s, faces: %v, model: %s)\n",
		taskID, mediaID, faceIDs, model)

	return taskID, nil
}

// GetTaskStatus returns the current status of a task
func GetTaskStatus(taskID string) (*SwapTask, error) {
	// Try database first
	ctx := context.Background()
	if repository.IsDBAvailable() {
		dbTask, err := repository.GetSwapTask(ctx, taskID)
		if err != nil {
			return nil, err
		}
		if dbTask != nil {
			return &SwapTask{
				ID:        dbTask.TaskID,
				MediaID:   dbTask.MediaID,
				FaceIDs:   dbTask.FaceIDs,
				Model:     dbTask.Model,
				Status:    dbTask.Status,
				ResultURL: dbTask.ResultURL.String,
				Error:     dbTask.ErrorMessage.String,
				CreatedAt: dbTask.CreatedAt,
				UpdatedAt: dbTask.UpdatedAt,
			}, nil
		}
	}

	// Fallback to in-memory
	taskMu.RLock()
	task, exists := taskStore[taskID]
	taskMu.RUnlock()

	if !exists {
		return nil, errors.New("task not found")
	}

	return task, nil
}

func simulateProcessing(taskID string) {
	// Simulate processing time (5-10 seconds)
	time.Sleep(5 * time.Second)

	resultURL := fmt.Sprintf("https://mock.deepswap.ai/results/%s.mp4", taskID)

	// Update database if available
	ctx := context.Background()
	if repository.IsDBAvailable() {
		repository.UpdateSwapTaskStatus(ctx, taskID, "completed", &resultURL, nil)
	}

	// Update in-memory
	taskMu.Lock()
	if task, exists := taskStore[taskID]; exists {
		task.Status = "completed"
		task.ResultURL = resultURL
		task.UpdatedAt = time.Now()
	}
	taskMu.Unlock()

	fmt.Printf("[MOCK] Task %s completed\n", taskID)
}

func generateMediaID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("media_%x", b)
}

func generateTaskID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("task_%x", b)
}
