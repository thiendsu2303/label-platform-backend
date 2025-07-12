package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioClient wraps the MinIO client
type MinioClient struct {
	client *minio.Client
	bucket string
}

// NewMinioClient creates a new MinIO client
func NewMinioClient() (*MinioClient, error) {
	useSSL, _ := strconv.ParseBool(os.Getenv("MINIO_USE_SSL"))

	client, err := minio.New(os.Getenv("MINIO_ENDPOINT"), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("MINIO_ACCESS_KEY"), os.Getenv("MINIO_SECRET_KEY"), ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	bucketName := os.Getenv("MINIO_BUCKET_NAME")

	// Check if bucket exists, if not create it
	exists, err := client.BucketExists(context.Background(), bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Printf("Created bucket: %s", bucketName)
	}

	log.Printf("Successfully connected to MinIO and verified bucket: %s", bucketName)

	return &MinioClient{
		client: client,
		bucket: bucketName,
	}, nil
}

// GetClient returns the underlying MinIO client
func (m *MinioClient) GetClient() *minio.Client {
	return m.client
}

// GetBucket returns the bucket name
func (m *MinioClient) GetBucket() string {
	return m.bucket
}
