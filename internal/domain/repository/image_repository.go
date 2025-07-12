package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/label-platform-backend/internal/domain/entity"
)

// ImageRepository defines the interface for image data operations
type ImageRepository interface {
	Create(ctx context.Context, image *entity.Image) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Image, error)
	GetAll(ctx context.Context) ([]*entity.Image, error)
	Update(ctx context.Context, image *entity.Image) error
	Delete(ctx context.Context, id uuid.UUID) error
}
