package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestImage_TableName(t *testing.T) {
	image := &Image{}
	assert.Equal(t, "images", image.TableName())
}

func TestImage_Creation(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	image := &Image{
		ID:               id,
		Name:             "test-image.png",
		MinioPath:        "test-path.png",
		GroundTruth:      map[string]any{"label": "button"},
		PredictedLabels:  map[string]any{"model1": map[string]any{"label": "button", "confidence": 0.95}},
		EvaluationScores: map[string]any{"accuracy": 0.92},
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	assert.Equal(t, id, image.ID)
	assert.Equal(t, "test-image.png", image.Name)
	assert.Equal(t, "test-path.png", image.MinioPath)
	assert.Equal(t, map[string]any{"label": "button"}, image.GroundTruth)
	assert.Equal(t, map[string]any{"model1": map[string]any{"label": "button", "confidence": 0.95}}, image.PredictedLabels)
	assert.Equal(t, map[string]any{"accuracy": 0.92}, image.EvaluationScores)
	assert.Equal(t, now, image.CreatedAt)
	assert.Equal(t, now, image.UpdatedAt)
}
