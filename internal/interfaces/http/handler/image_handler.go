package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/label-platform-backend/internal/domain/usecase"
)

// ImageHandler handles HTTP requests for images
type ImageHandler struct {
	imageUseCase usecase.ImageUseCase
}

// NewImageHandler creates a new image handler
func NewImageHandler(imageUseCase usecase.ImageUseCase) *ImageHandler {
	return &ImageHandler{
		imageUseCase: imageUseCase,
	}
}

// UploadImage handles image upload requests
func (h *ImageHandler) UploadImage(c *gin.Context) {
	// Check if file is present in the request
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "No image file provided. Please include a file with field name 'image'",
			"details": "Expected multipart/form-data with field 'image' containing the image file",
		})
		return
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         "Invalid file type. Please upload an image file (PNG, JPG, JPEG, etc.)",
			"received_type": contentType,
		})
		return
	}

	// Validate file size (optional - 10MB limit)
	if file.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "File too large. Maximum size is 10MB",
			"file_size": file.Size,
			"max_size":  10 * 1024 * 1024,
		})
		return
	}

	// Parse ground truth from form data
	groundTruthStr := c.PostForm("ground_truth")
	var groundTruth map[string]any
	if groundTruthStr != "" {
		if err := json.Unmarshal([]byte(groundTruthStr), &groundTruth); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid ground truth format. Please provide valid JSON",
				"details": err.Error(),
			})
			return
		}
	}

	// Upload image
	image, err := h.imageUseCase.UploadImage(c.Request.Context(), file, groundTruth)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to upload image",
			"details": err.Error(),
		})
		return
	}

	// Generate signed URL for the uploaded image
	signedURL, err := h.imageUseCase.GetImageURL(c.Request.Context(), image.MinioPath, time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate image URL",
			"details": err.Error(),
		})
		return
	}

	// Add signed URL to response
	response := gin.H{
		"id":                image.ID,
		"name":              image.Name,
		"minio_path":        image.MinioPath,
		"image_url":         signedURL,
		"ground_truth":      image.GroundTruth,
		"predicted_labels":  image.PredictedLabels,
		"evaluation_scores": image.EvaluationScores,
		"created_at":        image.CreatedAt,
		"updated_at":        image.UpdatedAt,
		"file_info": gin.H{
			"size":         file.Size,
			"content_type": contentType,
		},
	}

	c.JSON(http.StatusCreated, response)
}

// GetImageByID handles requests to get a specific image
func (h *ImageHandler) GetImageByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	image, err := h.imageUseCase.GetImageByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	// Generate signed URL for the image
	signedURL, err := h.imageUseCase.GetImageURL(c.Request.Context(), image.MinioPath, time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate image URL"})
		return
	}

	// Add signed URL to response
	response := gin.H{
		"id":                image.ID,
		"name":              image.Name,
		"minio_path":        image.MinioPath,
		"image_url":         signedURL,
		"ground_truth":      image.GroundTruth,
		"predicted_labels":  image.PredictedLabels,
		"evaluation_scores": image.EvaluationScores,
		"created_at":        image.CreatedAt,
		"updated_at":        image.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// GetAllImages handles requests to get all images
func (h *ImageHandler) GetAllImages(c *gin.Context) {
	images, err := h.imageUseCase.GetAllImages(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate signed URLs for all images
	var response []gin.H
	for _, image := range images {
		signedURL, err := h.imageUseCase.GetImageURL(c.Request.Context(), image.MinioPath, time.Hour)
		if err != nil {
			// Skip this image if URL generation fails
			continue
		}

		response = append(response, gin.H{
			"id":                image.ID,
			"name":              image.Name,
			"minio_path":        image.MinioPath,
			"image_url":         signedURL,
			"ground_truth":      image.GroundTruth,
			"predicted_labels":  image.PredictedLabels,
			"evaluation_scores": image.EvaluationScores,
			"created_at":        image.CreatedAt,
			"updated_at":        image.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

// UpdateImage handles requests to update image predictions
func (h *ImageHandler) UpdateImage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var request struct {
		PredictedLabels  map[string]any `json:"predicted_labels"`
		EvaluationScores map[string]any `json:"evaluation_scores"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	image, err := h.imageUseCase.UpdateImage(c.Request.Context(), id, request.PredictedLabels, request.EvaluationScores)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, image)
}

// DeleteImage handles requests to delete an image
func (h *ImageHandler) DeleteImage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = h.imageUseCase.DeleteImage(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Image deleted successfully"})
}

// GetImageURL handles requests to get a signed URL for an image
func (h *ImageHandler) GetImageURL(c *gin.Context) {
	imageIDStr := c.Param("id")
	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Get image from database
	image, err := h.imageUseCase.GetImageByID(c.Request.Context(), imageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	// Generate signed URL (expires in 1 hour)
	signedURL, err := h.imageUseCase.GetImageURL(c.Request.Context(), image.MinioPath, time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate signed URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"image_url":  signedURL,
		"expires_in": "1 hour",
	})
}
