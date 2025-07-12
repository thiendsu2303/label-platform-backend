.PHONY: help build run test clean deps docker-up docker-down

# Default target
help:
	@echo "Available commands:"
	@echo "  deps       - Download dependencies"
	@echo "  build      - Build the application"
	@echo "  run        - Run the application"
	@echo "  test       - Run tests"
	@echo "  clean      - Clean build artifacts"
	@echo "  docker-up  - Start PostgreSQL and MinIO containers"
	@echo "  docker-down- Stop PostgreSQL and MinIO containers"

# Download dependencies
deps:
	go mod download
	go mod tidy

# Build the application
build:
	go build -o bin/label-platform-backend cmd/main.go

# Run the application
run:
	go run cmd/main.go

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Start Docker services
docker-up:
	docker-compose up -d

# Stop Docker services
docker-down:
	docker-compose down

# Setup development environment
setup: docker-up deps
	@echo "Development environment setup complete!"
	@echo "PostgreSQL: localhost:5432"
	@echo "MinIO: localhost:9000 (API), localhost:9001 (Console)"
	@echo "Copy env.example to .env and run 'make run' to start the server" 