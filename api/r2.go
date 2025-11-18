package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type R2Client struct {
	client *minio.Client
	bucket string
}

func NewR2Client(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*R2Client, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	// ensure bucket exists (optional)
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, bucket)
	if err != nil {
		// some R2 setups return error for BucketExists even if bucket exists; ignore but log
		log.Printf("bucket exists check err: %v", err)
	}
	if !exists {
		if err := minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("create bucket: %w", err)
		}
	}

	return &R2Client{client: minioClient, bucket: bucket}, nil
}

// UploadReader uploads content from reader to objectName. Returns nil or error.
func (r *R2Client) UploadReader(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	_, err := r.client.PutObject(ctx, r.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

// PresignedGet generates a presigned URL valid for duration
func (r *R2Client) PresignedGet(ctx context.Context, objectName string, expires time.Duration) (string, error) {
	reqParams := make(map[string]string)
	u, err := r.client.PresignedGetObject(ctx, r.bucket, objectName, expires, reqParams)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
