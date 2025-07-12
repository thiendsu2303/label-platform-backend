package usecase

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/label-platform-backend/internal/domain/entity"
	"github.com/label-platform-backend/internal/domain/repository"
	"github.com/label-platform-backend/internal/infrastructure/storage"
	"github.com/minio/minio-go/v7"
)

// ImageUseCaseImpl implements the ImageUseCase interface
type ImageUseCaseImpl struct {
	imageRepo   repository.ImageRepository
	minioClient *storage.MinioClient
}

// NewImageUseCase creates a new image use case
func NewImageUseCase(imageRepo repository.ImageRepository, minioClient *storage.MinioClient) *ImageUseCaseImpl {
	return &ImageUseCaseImpl{
		imageRepo:   imageRepo,
		minioClient: minioClient,
	}
}

// UploadImage handles the upload of an image file and creates a new image
func (u *ImageUseCaseImpl) UploadImage(ctx context.Context, file *multipart.FileHeader, groundTruth map[string]any) (*entity.Image, error) {
	// Generate unique filename with format: screenshots/{uuid}-{original_filename}
	uuidStr := uuid.New().String()
	filename := fmt.Sprintf("screenshots/%s-%s", uuidStr, file.Filename)

	// Upload file to MinIO
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Upload to MinIO
	_, err = u.minioClient.GetClient().PutObject(ctx, u.minioClient.GetBucket(), filename, src, file.Size, minio.PutObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to MinIO: %w", err)
	}

	// Create image entity
	image := &entity.Image{
		ID:          uuid.MustParse(uuidStr),
		Name:        file.Filename,
		MinioPath:   filename,
		GroundTruth: groundTruth,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save to database
	err = u.imageRepo.Create(ctx, image)
	if err != nil {
		return nil, fmt.Errorf("failed to save image: %w", err)
	}

	return image, nil
}

// GetImageByID retrieves an image by its ID
func (u *ImageUseCaseImpl) GetImageByID(ctx context.Context, id uuid.UUID) (*entity.Image, error) {
	return u.imageRepo.GetByID(ctx, id)
}

// GetAllImages retrieves all images
func (u *ImageUseCaseImpl) GetAllImages(ctx context.Context) ([]*entity.Image, error) {
	return u.imageRepo.GetAll(ctx)
}

// UpdateImage updates an image with predicted labels and evaluation scores
func (u *ImageUseCaseImpl) UpdateImage(ctx context.Context, id uuid.UUID, predictedLabels map[string]any, evaluationScores map[string]any) (*entity.Image, error) {
	image, err := u.imageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	image.PredictedLabels = predictedLabels
	image.EvaluationScores = evaluationScores
	image.UpdatedAt = time.Now()

	err = u.imageRepo.Update(ctx, image)
	if err != nil {
		return nil, fmt.Errorf("failed to update image: %w", err)
	}

	return image, nil
}

// DeleteImage removes an image and its associated file
func (u *ImageUseCaseImpl) DeleteImage(ctx context.Context, id uuid.UUID) error {
	image, err := u.imageRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get image: %w", err)
	}

	// Delete file from MinIO
	err = u.minioClient.GetClient().RemoveObject(ctx, u.minioClient.GetBucket(), image.MinioPath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file from MinIO: %w", err)
	}

	// Delete from database
	return u.imageRepo.Delete(ctx, id)
}

// GetImageURL generates a signed URL for accessing the image
func (u *ImageUseCaseImpl) GetImageURL(ctx context.Context, minioPath string, expiry time.Duration) (string, error) {
	url, err := u.minioClient.GetClient().PresignedGetObject(ctx, u.minioClient.GetBucket(), minioPath, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}
	return url.String(), nil
}
