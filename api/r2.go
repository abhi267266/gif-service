package main

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"strconv"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// R2Client wraps the MinIO client for Cloudflare R2 operations
type R2Client struct {
	client     *minio.Client
	bucketName string
}

// NewR2Client creates and initializes a new R2 client
func NewR2Client() (*R2Client, error) {
	endpoint := os.Getenv("R2_ENDPOINT")
	accessKey := os.Getenv("R2_ACCESS_KEY")
	secretKey := os.Getenv("R2_SECRET_KEY")
	bucketName := os.Getenv("R2_BUCKET")
	useSSL := os.Getenv("R2_USE_SSL")

	// Validate required environment variables
	if endpoint == "" || accessKey == "" || secretKey == "" || bucketName == "" {
		return nil, fmt.Errorf("missing required R2 configuration (R2_ENDPOINT, R2_ACCESS_KEY, R2_SECRET_KEY, R2_BUCKET)")
	}

	// Parse SSL flag
	ssl := true
	if useSSL != "" {
		var err error
		ssl, err = strconv.ParseBool(useSSL)
		if err != nil {
			log.Printf("Invalid R2_USE_SSL value, defaulting to true")
		}
	}

	// Remove https:// or http:// from endpoint
	endpoint = removeProtocol(endpoint)

	// Initialize MinIO client
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: ssl,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	log.Printf("R2 client initialized successfully for bucket: %s", bucketName)

	return &R2Client{
		client:     minioClient,
		bucketName: bucketName,
	}, nil
}

// UploadFile uploads a file to R2 bucket
func (r *R2Client) UploadFile(file *multipart.FileHeader) (*UploadResult, error) {
	// Open the file
	fileContent, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer fileContent.Close()

	// Determine content type
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload file to R2
	objectName := file.Filename
	info, err := r.client.PutObject(
		context.Background(),
		r.bucketName,
		objectName,
		fileContent,
		file.Size,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to R2: %w", err)
	}

	return &UploadResult{
		Filename: objectName,
		Size:     info.Size,
		Bucket:   r.bucketName,
		ETag:     info.ETag,
	}, nil
}

// UploadResult contains information about the uploaded file
type UploadResult struct {
	Filename string
	Size     int64
	Bucket   string
	ETag     string
}

// removeProtocol removes http:// or https:// prefix from endpoint URL
func removeProtocol(endpoint string) string {
	if len(endpoint) > 8 && endpoint[:8] == "https://" {
		return endpoint[8:]
	}
	if len(endpoint) > 7 && endpoint[:7] == "http://" {
		return endpoint[7:]
	}
	return endpoint
}
