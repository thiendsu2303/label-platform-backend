package entity

import (
	"time"

	"github.com/google/uuid"
)

// Image represents the core domain entity for images
type Image struct {
	ID               uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name             string         `json:"name" gorm:"type:text;not null"`
	MinioPath        string         `json:"minio_path" gorm:"type:text;not null"`
	GroundTruth      map[string]any `json:"ground_truth" gorm:"type:jsonb"`
	PredictedLabels  map[string]any `json:"predicted_labels" gorm:"type:jsonb"`
	EvaluationScores map[string]any `json:"evaluation_scores" gorm:"type:jsonb"`
	CreatedAt        time.Time      `json:"created_at" gorm:"default:now()"`
	UpdatedAt        time.Time      `json:"updated_at" gorm:"default:now()"`
}

// TableName specifies the table name for GORM
func (Image) TableName() string {
	return "images"
}
