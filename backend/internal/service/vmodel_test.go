package service

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestVModelCredits(t *testing.T) {
	token := os.Getenv("VMODEL_API_TOKEN")
	if token == "" {
		t.Skip("VMODEL_API_TOKEN not set")
	}

	client := GetVModelClient()
	credits, err := client.GetCredits(context.Background())
	if err != nil {
		t.Fatalf("GetCredits failed: %v", err)
	}

	t.Logf("VModel credits: %.2f", credits)
	if credits <= 0 {
		t.Error("Expected positive credits balance")
	}
}

func TestVModelDetectFaces(t *testing.T) {
	token := os.Getenv("VMODEL_API_TOKEN")
	if token == "" {
		t.Skip("VMODEL_API_TOKEN not set")
	}

	// Use official example video
	testVideoURL := "https://vmodel.ai/data/model/remaker/video-face-detect/tmp2ukv7myu.mp4"

	client := GetVModelClient()
	result, err := client.DetectFaces(context.Background(), testVideoURL)
	if err != nil {
		t.Fatalf("DetectFaces failed: %v", err)
	}

	t.Logf("DetectID: %s", result.DetectID)
	t.Logf("Faces found: %d", len(result.Faces))

	if result.DetectID == "" {
		t.Error("Expected non-empty DetectID")
	}

	for i, face := range result.Faces {
		t.Logf("  Face %d: ID=%d, Thumbnail=%s", i, face.ID, face.Link)
	}
}

func TestVModelCreateSwapTask(t *testing.T) {
	token := os.Getenv("VMODEL_API_TOKEN")
	if token == "" {
		t.Skip("VMODEL_API_TOKEN not set")
	}

	// First detect faces
	testVideoURL := "https://vmodel.ai/data/model/remaker/video-face-detect/tmp2ukv7myu.mp4"
	testFaceURL := "https://vmodel.ai/data/model/remaker/video-face-detect/tmp2ukv7myu.mp4" // Use same for test

	client := GetVModelClient()
	detectResult, err := client.DetectFaces(context.Background(), testVideoURL)
	if err != nil {
		t.Fatalf("DetectFaces failed: %v", err)
	}

	if len(detectResult.Faces) == 0 {
		t.Skip("No faces detected, skipping swap test")
	}

	t.Logf("Using DetectID: %s", detectResult.DetectID)
	t.Logf("Using FaceID: %d", detectResult.Faces[0].ID)

	// Create swap task
	faceSwaps := []VModelFaceSwapPair{
		{
			FaceID: detectResult.Faces[0].ID,
			Target: testFaceURL,
		},
	}

	swapResult, err := client.CreateSwapTask(context.Background(), detectResult.DetectID, faceSwaps, true)
	if err != nil {
		t.Fatalf("CreateSwapTask failed: %v", err)
	}

	t.Logf("Swap TaskID: %s", swapResult.TaskID)
	t.Logf("Swap Status: %s", swapResult.Status)

	if swapResult.TaskID == "" {
		t.Error("Expected non-empty TaskID")
	}
}

func TestVModelGetTaskStatus(t *testing.T) {
	token := os.Getenv("VMODEL_API_TOKEN")
	if token == "" {
		t.Skip("VMODEL_API_TOKEN not set")
	}

	// Use the task ID from previous test (you may need to update this)
	taskID := "dey1bnqvw661w0xlsf" // Task ID from the curl test

	client := GetVModelClient()
	result, err := client.GetTaskStatus(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetTaskStatus failed: %v", err)
	}

	t.Logf("TaskID: %s", result.TaskID)
	t.Logf("Status: %s", result.Status)
	t.Logf("ResultURL: %s", result.ResultURL)
	t.Logf("Error: %s", result.Error)
}

func TestVModelFullFlow(t *testing.T) {
	token := os.Getenv("VMODEL_API_TOKEN")
	if token == "" {
		t.Skip("VMODEL_API_TOKEN not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	client := GetVModelClient()

	// Step 1: Check credits
	credits, err := client.GetCredits(ctx)
	if err != nil {
		t.Fatalf("GetCredits failed: %v", err)
	}
	t.Logf("Credits available: %.2f", credits)

	// Step 2: Detect faces
	testVideoURL := "https://vmodel.ai/data/model/remaker/video-face-detect/tmp2ukv7myu.mp4"
	detectResult, err := client.DetectFaces(ctx, testVideoURL)
	if err != nil {
		t.Fatalf("DetectFaces failed: %v", err)
	}
	t.Logf("Detected %d faces, DetectID: %s", len(detectResult.Faces), detectResult.DetectID)

	if len(detectResult.Faces) == 0 {
		t.Fatal("No faces detected")
	}

	t.Log("Full flow test passed (swap task creation skipped to save credits)")
}
