package entity

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

func TestImage_TableName(t *testing.T) {
	image := &Image{}
	assert.Equal(t, "images", image.TableName())
}

func TestImage_Creation(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	// Create test JSON data
	groundTruthJSON, _ := json.Marshal(map[string]any{"label": "button"})
	predictedLabelsJSON, _ := json.Marshal(map[string]any{"model1": map[string]any{"label": "button", "confidence": 0.95}})
	evaluationScoresJSON, _ := json.Marshal(map[string]any{"accuracy": 0.92})

	image := &Image{
		ID:               id,
		Name:             "test-image.png",
		MinioPath:        "test-path.png",
		GroundTruth:      datatypes.JSON(groundTruthJSON),
		PredictedLabels:  datatypes.JSON(predictedLabelsJSON),
		EvaluationScores: datatypes.JSON(evaluationScoresJSON),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	assert.Equal(t, id, image.ID)
	assert.Equal(t, "test-image.png", image.Name)
	assert.Equal(t, "test-path.png", image.MinioPath)
	assert.Equal(t, datatypes.JSON(groundTruthJSON), image.GroundTruth)
	assert.Equal(t, datatypes.JSON(predictedLabelsJSON), image.PredictedLabels)
	assert.Equal(t, datatypes.JSON(evaluationScoresJSON), image.EvaluationScores)
	assert.Equal(t, now, image.CreatedAt)
	assert.Equal(t, now, image.UpdatedAt)
}
