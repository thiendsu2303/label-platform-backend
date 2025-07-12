package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/label-platform-backend/internal/interfaces/http/handler"
)

// SetupRouter configures the HTTP router with all endpoints
func SetupRouter(imageHandler *handler.ImageHandler) *gin.Engine {
	router := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	router.Use(cors.New(config))

	// API routes
	api := router.Group("/api/v1")
	{
		// Image routes
		images := api.Group("/images")
		{
			images.POST("/upload", imageHandler.UploadImage)
			images.GET("/", imageHandler.GetAllImages)
			images.GET("/:id", imageHandler.GetImageByID)
			images.GET("/:id/url", imageHandler.GetImageURL)
			images.PUT("/:id", imageHandler.UpdateImage)
			images.PUT("/:id/ground-truth", imageHandler.UpdateGroundTruth)
			images.DELETE("/:id", imageHandler.DeleteImage)
		}
	}

	return router
}
