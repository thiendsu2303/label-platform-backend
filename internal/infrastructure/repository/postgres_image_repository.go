package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/label-platform-backend/internal/domain/entity"
	"github.com/label-platform-backend/internal/domain/repository"
	"gorm.io/gorm"
)

// PostgresImageRepository implements the ImageRepository interface
type PostgresImageRepository struct {
	db *gorm.DB
}

// NewPostgresImageRepository creates a new PostgreSQL repository
func NewPostgresImageRepository(db *gorm.DB) repository.ImageRepository {
	return &PostgresImageRepository{db: db}
}

// Create saves a new image to the database
func (r *PostgresImageRepository) Create(ctx context.Context, image *entity.Image) error {
	return r.db.WithContext(ctx).Create(image).Error
}

// GetByID retrieves an image by its ID
func (r *PostgresImageRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Image, error) {
	var image entity.Image
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&image).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

// GetAll retrieves all images
func (r *PostgresImageRepository) GetAll(ctx context.Context) ([]*entity.Image, error) {
	var images []*entity.Image
	err := r.db.WithContext(ctx).Find(&images).Error
	if err != nil {
		return nil, err
	}
	return images, nil
}

// Update updates an existing image
func (r *PostgresImageRepository) Update(ctx context.Context, image *entity.Image) error {
	return r.db.WithContext(ctx).Save(image).Error
}

// Delete removes an image by its ID
func (r *PostgresImageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Image{}).Error
}
