package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/abhi267266/gif-service/api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var r2Client *R2Client

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Initialize Database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	if err := models.InitDB(dbURL); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize R2 client
	var err error
	r2Client, err = NewR2Client()
	if err != nil {
		log.Fatalf("Failed to initialize R2 client: %v", err)
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit: 100 * 1024 * 1024, // 100MB max file size
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Routes
	app.Get("/", homeHandler)
	app.Post("/upload", uploadHandler)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	log.Printf("Server starting on port %s...", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func homeHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "File Upload API - Cloudflare R2",
		"endpoints": fiber.Map{
			"POST /upload": "Upload a video file to R2",
		},
	})
}

func uploadHandler(c *fiber.Ctx) error {
	// Get the file from the request
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file uploaded",
		})
	}

	// Validate file type (Video only)
	contentType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "video/") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Only video files are allowed",
		})
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	newFilename := uuid.New().String() + ext

	// Update file header to use new filename for upload
	originalFilename := file.Filename
	file.Filename = newFilename

	// Upload file to R2
	result, err := r2Client.UploadFile(file)
	if err != nil {
		log.Printf("Upload error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to upload file to R2",
		})
	}

	// Save metadata to Database
	video := &models.Video{
		Filename:     result.Filename,
		OriginalName: originalFilename,
		Size:         result.Size,
		Bucket:       result.Bucket,
		URL:          fmt.Sprintf("%s/%s", os.Getenv("R2_PUBLIC_URL"), result.Filename), // Assuming public URL construction
	}

	// If R2_PUBLIC_URL is not set, use a placeholder or construct from endpoint
	if os.Getenv("R2_PUBLIC_URL") == "" {
		video.URL = fmt.Sprintf("https://%s.%s/%s", result.Bucket, "r2.cloudflarestorage.com", result.Filename)
	}

	if err := models.CreateVideo(video); err != nil {
		log.Printf("Database error: %v", err)
		// Note: File is uploaded but DB failed. In a real app, might want to delete file or retry.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "File uploaded but failed to save metadata",
		})
	}

	return c.JSON(fiber.Map{
		"message":       "File uploaded successfully",
		"id":            video.ID,
		"filename":      video.Filename,
		"original_name": video.OriginalName,
		"size":          video.Size,
		"bucket":        video.Bucket,
		"url":           video.URL,
	})
}
