package handler

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/label-platform-backend/internal/domain/usecase"
	"github.com/label-platform-backend/internal/infrastructure/redis"
	"github.com/label-platform-backend/internal/infrastructure/storage"
	"github.com/minio/minio-go/v7"
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

	// Convert datatypes.JSON to map for response
	var groundTruthMap, predictedLabelsMap, evaluationScoresMap map[string]any

	if image.GroundTruth != nil {
		json.Unmarshal(image.GroundTruth, &groundTruthMap)
	}
	if image.PredictedLabels != nil {
		json.Unmarshal(image.PredictedLabels, &predictedLabelsMap)
	}
	if image.EvaluationScores != nil {
		json.Unmarshal(image.EvaluationScores, &evaluationScoresMap)
	}

	// Add signed URL to response
	response := gin.H{
		"id":                image.ID,
		"name":              image.Name,
		"minio_path":        image.MinioPath,
		"image_url":         signedURL,
		"ground_truth":      groundTruthMap,
		"predicted_labels":  predictedLabelsMap,
		"evaluation_scores": evaluationScoresMap,
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate image URL",
			"details": err.Error(),
		})
		return
	}

	// Convert datatypes.JSON to map for response
	var groundTruthMap, predictedLabelsMap, evaluationScoresMap map[string]any

	if image.GroundTruth != nil {
		json.Unmarshal(image.GroundTruth, &groundTruthMap)
	}
	if image.PredictedLabels != nil {
		json.Unmarshal(image.PredictedLabels, &predictedLabelsMap)
	}
	if image.EvaluationScores != nil {
		json.Unmarshal(image.EvaluationScores, &evaluationScoresMap)
	}

	// Add signed URL to response
	response := gin.H{
		"id":                image.ID,
		"name":              image.Name,
		"minio_path":        image.MinioPath,
		"image_url":         signedURL,
		"ground_truth":      groundTruthMap,
		"predicted_labels":  predictedLabelsMap,
		"evaluation_scores": evaluationScoresMap,
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

		// Convert datatypes.JSON to map for response
		var groundTruthMap, predictedLabelsMap, evaluationScoresMap map[string]any

		if image.GroundTruth != nil {
			json.Unmarshal(image.GroundTruth, &groundTruthMap)
		}
		if image.PredictedLabels != nil {
			json.Unmarshal(image.PredictedLabels, &predictedLabelsMap)
		}
		if image.EvaluationScores != nil {
			json.Unmarshal(image.EvaluationScores, &evaluationScoresMap)
		}

		response = append(response, gin.H{
			"id":                image.ID,
			"name":              image.Name,
			"minio_path":        image.MinioPath,
			"image_url":         signedURL,
			"ground_truth":      groundTruthMap,
			"predicted_labels":  predictedLabelsMap,
			"evaluation_scores": evaluationScoresMap,
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

// UpdateGroundTruth handles requests to update image ground truth
func (h *ImageHandler) UpdateGroundTruth(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var request struct {
		GroundTruth map[string]any `json:"ground_truth"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	image, err := h.imageUseCase.UpdateGroundTruth(c.Request.Context(), id, request.GroundTruth)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate signed URL for the image
	signedURL, err := h.imageUseCase.GetImageURL(c.Request.Context(), image.MinioPath, time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate image URL",
			"details": err.Error(),
		})
		return
	}

	// Convert datatypes.JSON to map for response
	var groundTruthMap, predictedLabelsMap, evaluationScoresMap map[string]any

	if image.GroundTruth != nil {
		json.Unmarshal(image.GroundTruth, &groundTruthMap)
	}
	if image.PredictedLabels != nil {
		json.Unmarshal(image.PredictedLabels, &predictedLabelsMap)
	}
	if image.EvaluationScores != nil {
		json.Unmarshal(image.EvaluationScores, &evaluationScoresMap)
	}

	// Add signed URL to response
	response := gin.H{
		"id":                image.ID,
		"name":              image.Name,
		"minio_path":        image.MinioPath,
		"image_url":         signedURL,
		"ground_truth":      groundTruthMap,
		"predicted_labels":  predictedLabelsMap,
		"evaluation_scores": evaluationScoresMap,
		"created_at":        image.CreatedAt,
		"updated_at":        image.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// PredictImage handles GET /api/v1/images/:id/predict
func (h *ImageHandler) PredictImage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID format"})
		return
	}

	// Rate limit: chỉ cho phép mỗi ảnh predict 1 lần mỗi 5 phút
	lockKey := "predict-lock:" + id.String()
	ctx := c.Request.Context()
	ttl, err := redis.RedisClient.TTL(ctx, lockKey).Result()
	if err == nil && ttl > 0 {
		c.JSON(429, gin.H{
			"error":               "Rate limited. Please wait before retrying.",
			"retry_after_seconds": int(ttl.Seconds()),
		})
		return
	}
	// Dùng SetNX để đảm bảo chỉ set khi chưa có key, và luôn có TTL
	ok, err := redis.RedisClient.SetNX(ctx, lockKey, "1", 5*time.Minute).Result()
	if err != nil {
		c.JSON(500, gin.H{"error": "Redis error", "details": err.Error()})
		return
	}
	if !ok {
		// Nếu key đã tồn tại (race condition), trả về rate limit luôn
		ttl, _ := redis.RedisClient.TTL(ctx, lockKey).Result()
		c.JSON(429, gin.H{
			"error":               "Rate limited. Please wait before retrying.",
			"retry_after_seconds": int(ttl.Seconds()),
		})
		return
	}

	// Lấy thông tin ảnh từ Postgres
	image, err := h.imageUseCase.GetImageByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Image not found"})
		return
	}

	// Lấy file ảnh từ MinIO
	minioClient := h.imageUseCase.(interface{ GetMinioClient() *storage.MinioClient }).GetMinioClient()
	obj, err := minioClient.GetClient().GetObject(c.Request.Context(), minioClient.GetBucket(), image.MinioPath, minio.GetObjectOptions{})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get image from MinIO"})
		return
	}
	defer obj.Close()
	imgBytes, err := io.ReadAll(obj)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to read image data"})
		return
	}

	// Encode base64
	imgBase64 := base64.StdEncoding.EncodeToString(imgBytes)

	// Tạo payload
	payload := map[string]any{
		"id":           image.ID.String(),
		"image_base64": imgBase64,
	}
	payloadJSON, _ := json.Marshal(payload)

	ctx = c.Request.Context()
	redis.RedisClient.RPush(ctx, redis.QueueGPT, payloadJSON)
	redis.RedisClient.RPush(ctx, redis.QueueClaude, payloadJSON)
	redis.RedisClient.RPush(ctx, redis.QueueGemini, payloadJSON)

	c.JSON(200, gin.H{
		"message": "Image pushed to model queues",
		"id":      image.ID,
	})
}
