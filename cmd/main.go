package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/label-platform-backend/internal/application/usecase"
	"github.com/label-platform-backend/internal/domain/entity"
	"github.com/label-platform-backend/internal/infrastructure/database"
	"github.com/label-platform-backend/internal/infrastructure/repository"
	"github.com/label-platform-backend/internal/infrastructure/storage"
	"github.com/label-platform-backend/internal/interfaces/http/handler"
	"github.com/label-platform-backend/internal/interfaces/http/router"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize database
	db, err := database.NewPostgresConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate database schema
	if err := db.AutoMigrate(&entity.Image{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize MinIO client
	minioClient, err := storage.NewMinioClient()
	if err != nil {
		log.Fatalf("Failed to connect to MinIO: %v", err)
	}

	// Initialize repositories
	imageRepo := repository.NewPostgresImageRepository(db)

	// Initialize use cases
	imageUseCase := usecase.NewImageUseCase(imageRepo, minioClient)

	// Initialize handlers
	imageHandler := handler.NewImageHandler(imageUseCase)

	// Setup router
	router := router.SetupRouter(imageHandler)

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
