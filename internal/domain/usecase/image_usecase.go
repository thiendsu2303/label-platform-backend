package usecase

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/label-platform-backend/internal/domain/entity"
)

// ImageUseCase defines the interface for image business logic
type ImageUseCase interface {
	UploadImage(ctx context.Context, file *multipart.FileHeader, groundTruth map[string]any) (*entity.Image, error)
	GetImageByID(ctx context.Context, id uuid.UUID) (*entity.Image, error)
	GetAllImages(ctx context.Context) ([]*entity.Image, error)
	UpdateImage(ctx context.Context, id uuid.UUID, predictedLabels map[string]any, evaluationScores map[string]any) (*entity.Image, error)
	DeleteImage(ctx context.Context, id uuid.UUID) error
	GetImageURL(ctx context.Context, minioPath string, expiry time.Duration) (string, error)
}
