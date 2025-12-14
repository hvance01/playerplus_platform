package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"playplus_platform/internal/config"
)

// StorageService handles file storage operations
type StorageService struct {
	cfg         *config.Config
	minioClient *minio.Client
	localDir    string
	bucketName  string
}

var (
	storageService *StorageService
	storageOnce    sync.Once
)

// GetStorageService returns the singleton storage service
func GetStorageService() *StorageService {
	storageOnce.Do(func() {
		cfg := config.Get()

		localDir := os.Getenv("LOCAL_STORAGE_DIR")
		if localDir == "" {
			localDir = "./uploads"
		}
		// Ensure directory exists
		os.MkdirAll(localDir, 0755)

		storageService = &StorageService{
			cfg:        cfg,
			localDir:   localDir,
			bucketName: cfg.StorageBucket,
		}

		// Initialize MinIO client if configured
		if cfg.IsStorageConfigured() {
			if err := storageService.initMinioClient(); err != nil {
				log.Printf("Warning: Failed to init MinIO client: %v, using local storage", err)
			}
		}
	})
	return storageService
}

// initMinioClient initializes the MinIO client
func (s *StorageService) initMinioClient() error {
	// Parse endpoint - remove protocol prefix
	endpoint := s.cfg.StorageEndpoint
	useSSL := true

	if strings.HasPrefix(endpoint, "https://") {
		endpoint = strings.TrimPrefix(endpoint, "https://")
		useSSL = true
	} else if strings.HasPrefix(endpoint, "http://") {
		endpoint = strings.TrimPrefix(endpoint, "http://")
		useSSL = false
	}

	// Remove trailing slash
	endpoint = strings.TrimSuffix(endpoint, "/")

	// Create MinIO client
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.cfg.StorageAccessKey, s.cfg.StorageSecretKey, ""),
		Secure: useSSL,
		Region: s.cfg.StorageRegion,
	})
	if err != nil {
		return fmt.Errorf("create minio client: %w", err)
	}

	s.minioClient = client

	// Ensure bucket exists
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return fmt.Errorf("check bucket exists: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{
			Region: s.cfg.StorageRegion,
		})
		if err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}

		// Set bucket policy to allow public read
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}]
		}`, s.bucketName)

		err = client.SetBucketPolicy(ctx, s.bucketName, policy)
		if err != nil {
			log.Printf("Warning: Failed to set bucket policy: %v", err)
		}

		log.Printf("Created bucket: %s", s.bucketName)
	}

	return nil
}

// GenerateKey generates a unique storage key for a file
func (s *StorageService) GenerateKey(prefix, filename string) string {
	b := make([]byte, 8)
	rand.Read(b)
	ext := filepath.Ext(filename)
	return fmt.Sprintf("%s/%s%s", prefix, hex.EncodeToString(b), ext)
}

// Upload uploads a file and returns its public URL
func (s *StorageService) Upload(ctx context.Context, key string, file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer src.Close()

	content, err := io.ReadAll(src)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return s.UploadBytes(ctx, key, content, contentType)
}

// UploadBytes uploads raw bytes and returns the public URL
func (s *StorageService) UploadBytes(ctx context.Context, key string, content []byte, contentType string) (string, error) {
	// If MinIO client configured, use it
	if s.minioClient != nil {
		return s.uploadToMinio(ctx, key, content, contentType)
	}

	// Fallback to local storage
	return s.uploadToLocal(key, content)
}

// uploadToMinio uploads to MinIO/S3-compatible storage
func (s *StorageService) uploadToMinio(ctx context.Context, key string, content []byte, contentType string) (string, error) {
	reader := bytes.NewReader(content)

	_, err := s.minioClient.PutObject(ctx, s.bucketName, key, reader, int64(len(content)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("minio upload: %w", err)
	}

	// Return the public URL
	return s.GetPublicURL(key), nil
}

// uploadToLocal saves file to local filesystem (development fallback)
func (s *StorageService) uploadToLocal(key string, content []byte) (string, error) {
	filePath := filepath.Join(s.localDir, key)

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	// Return local URL (served by the app)
	return fmt.Sprintf("/uploads/%s", key), nil
}

// UploadFromURL downloads a file from URL and uploads to storage
func (s *StorageService) UploadFromURL(ctx context.Context, sourceURL, key string) (string, error) {
	// Use a simple HTTP client for downloading
	resp, err := s.cfg.HTTPClient().Get(sourceURL)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("download error: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read download: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return s.UploadBytes(ctx, key, content, contentType)
}

// GetLocalPath returns the local filesystem path for a key
func (s *StorageService) GetLocalPath(key string) string {
	// Remove leading /uploads/ if present
	key = strings.TrimPrefix(key, "/uploads/")
	return filepath.Join(s.localDir, key)
}

// Delete removes a file from storage
func (s *StorageService) Delete(ctx context.Context, key string) error {
	if s.minioClient != nil {
		return s.minioClient.RemoveObject(ctx, s.bucketName, key, minio.RemoveObjectOptions{})
	}

	// Local delete
	return os.Remove(s.GetLocalPath(key))
}

// IsConfigured returns true if storage is properly configured
func (s *StorageService) IsConfigured() bool {
	return s.minioClient != nil
}

// GetPublicURL returns the public URL for a stored file
func (s *StorageService) GetPublicURL(key string) string {
	if s.minioClient != nil {
		// Use public URL format
		publicURL := s.cfg.StoragePublicURL
		if publicURL != "" {
			return fmt.Sprintf("%s/%s/%s", strings.TrimSuffix(publicURL, "/"), s.bucketName, key)
		}
		// Fallback to endpoint
		endpoint := s.cfg.StorageEndpoint
		if !strings.HasPrefix(endpoint, "http") {
			endpoint = "https://" + endpoint
		}
		return fmt.Sprintf("%s/%s/%s", strings.TrimSuffix(endpoint, "/"), s.bucketName, key)
	}
	return fmt.Sprintf("/uploads/%s", key)
}

// GetPresignedURL returns a presigned URL for temporary access
func (s *StorageService) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if s.minioClient == nil {
		return s.GetPublicURL(key), nil
	}

	url, err := s.minioClient.PresignedGetObject(ctx, s.bucketName, key, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("generate presigned url: %w", err)
	}

	return url.String(), nil
}
