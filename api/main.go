package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

var r2Client *R2Client

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
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
			"POST /upload": "Upload a file to R2",
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

	// Upload file to R2
	result, err := r2Client.UploadFile(file)
	if err != nil {
		log.Printf("Upload error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to upload file to R2",
		})
	}

	return c.JSON(fiber.Map{
		"message":  "File uploaded successfully",
		"filename": result.Filename,
		"size":     result.Size,
		"bucket":   result.Bucket,
		"etag":     result.ETag,
	})
}
