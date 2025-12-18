package config

import (
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	cfg        *Config
	once       sync.Once
	httpClient *http.Client
)

type Config struct {
	// Server
	Port string

	// Database
	DatabaseURL string

	// VModel API
	VModelAPIToken string
	VModelBaseURL  string

	// Storage (MinIO / S3)
	StorageBucket    string
	StorageEndpoint  string
	StorageRegion    string
	StorageAccessKey string
	StorageSecretKey string
	StoragePublicURL  string // Public URL for accessing stored files (CDN)
	StorageDirectURL  string // Direct URL for external API access (bypassing CDN)

	// Resend (Email)
	ResendAPIKey string
}

func Get() *Config {
	once.Do(func() {
		cfg = &Config{
			// Server
			Port: getEnv("PORT", "8080"),

			// Database
			DatabaseURL: os.Getenv("DATABASE_URL"),

			// VModel API
			VModelAPIToken: getEnv("VMODEL_API_TOKEN", ""),
			VModelBaseURL:  getEnv("VMODEL_BASE_URL", "https://api.vmodel.ai"),

			// Storage
			StorageBucket:    getEnv("BUCKET_NAME", "playerplus-media"),
			StorageEndpoint:  getEnv("MINIO_PUBLIC_ENDPOINT", getEnv("AWS_ENDPOINT_URL", "")),
			StorageRegion:    getEnv("AWS_REGION", "us-east-1"),
			StorageAccessKey: getEnv("MINIO_ROOT_USER", os.Getenv("AWS_ACCESS_KEY_ID")),
			StorageSecretKey: getEnv("MINIO_ROOT_PASSWORD", os.Getenv("AWS_SECRET_ACCESS_KEY")),
			StoragePublicURL:  getEnv("STORAGE_PUBLIC_URL", ""),
			StorageDirectURL:  getEnv("STORAGE_DIRECT_URL", ""),

			// Resend
			ResendAPIKey: os.Getenv("RESEND_API_KEY"),
		}
	})
	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsStorageConfigured checks if storage is properly configured
func (c *Config) IsStorageConfigured() bool {
	return c.StorageAccessKey != "" && c.StorageSecretKey != ""
}

// IsVModelConfigured checks if VModel API is properly configured
func (c *Config) IsVModelConfigured() bool {
	return c.VModelAPIToken != ""
}

// IsFaceSwapConfigured checks if face swap API is configured
func (c *Config) IsFaceSwapConfigured() bool {
	return c.IsVModelConfigured()
}

// HTTPClient returns a shared HTTP client for making requests
func (c *Config) HTTPClient() *http.Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 5 * time.Minute, // 5 min for large file downloads
		}
	}
	return httpClient
}
