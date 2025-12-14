package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"playplus_platform/internal/handler"
	"playplus_platform/internal/repository"
)

func main() {
	// Load .env file if exists (for local development)
	if err := godotenv.Load(); err != nil {
		log.Println("[INFO] No .env file found, using environment variables")
	}

	// Initialize database
	if err := repository.InitDB(); err != nil {
		log.Printf("[WARN] Database initialization failed: %v", err)
		log.Println("[INFO] Running without database persistence")
	}
	defer repository.CloseDB()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := handler.SetupRouter()

	log.Printf("Server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
