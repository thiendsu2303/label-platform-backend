# Label Platform Backend

A complete Golang backend server built with Gin and GORM, following Clean Architecture principles. This backend provides REST APIs for uploading UI design screenshots, storing images on MinIO, and saving annotation metadata in PostgreSQL.

## Features

- **Clean Architecture**: Follows domain-driven design principles with clear separation of concerns
- **Image Upload**: Upload UI design screenshots with metadata
- **MinIO Storage**: Secure object storage for uploaded images
- **PostgreSQL Database**: Reliable data storage with JSONB support for flexible metadata
- **RESTful API**: Complete CRUD operations for images
- **CORS Support**: Cross-origin resource sharing enabled
- **Graceful Shutdown**: Proper server shutdown handling

## Architecture

```
├── cmd/main.go                    # Application entry point
├── internal/
│   ├── domain/                    # Domain layer (entities, interfaces)
│   │   ├── entity/
│   │   ├── repository/
│   │   └── usecase/
│   ├── application/               # Application layer (use case implementations)
│   │   └── usecase/
│   ├── infrastructure/            # Infrastructure layer (external services)
│   │   ├── database/
│   │   ├── repository/
│   │   └── storage/
│   └── interfaces/                # Interface layer (HTTP handlers, routers)
│       └── http/
│           ├── handler/
│           └── router/
```

## Database Schema

```sql
CREATE TABLE images (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  minio_path TEXT NOT NULL,
  ground_truth JSONB,
  predicted_labels JSONB,
  evaluation_scores JSONB,
  created_at TIMESTAMP DEFAULT now(),
  updated_at TIMESTAMP DEFAULT now()
);
```

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL 15
- MinIO

## Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd label-platform-backend
   ```

2. **Setup environment**
   ```bash
   cp env.example .env
   # Edit .env with your configuration
   ```

3. **Start dependencies**
   ```bash
   docker network create label-platform-network
   make docker-up
   ```

4. **Install dependencies**
   ```bash
   make deps
   ```

5. **Run the application**
   ```bash
   make run
   ```

The server will start on `http://localhost:8080`

## API Endpoints

### Upload Image
```
POST /api/v1/images/upload
Content-Type: multipart/form-data

Form Data:
- image: File (required) - The image file to upload
- ground_truth: JSON string (optional) - Ground truth labels in JSON format

Features:
- Supports common image formats (PNG, JPG, JPEG, etc.)
- File size limit: 10MB
- Stores images in MinIO with path format: screenshots/{uuid}-{original_filename}
- Validates file type and size
- Returns signed URL for immediate access
```

**Example Request:**
```javascript
const formData = new FormData();
formData.append('image', fileInput.files[0]);
formData.append('ground_truth', JSON.stringify({
  "elements": [
    {"type": "button", "text": "Submit", "position": {"x": 100, "y": 200}},
    {"type": "input", "placeholder": "Enter text", "position": {"x": 100, "y": 150}}
  ]
}));

fetch('/api/v1/images/upload', {
  method: 'POST',
  body: formData
});
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "ui-design.png",
  "minio_path": "screenshots/550e8400-e29b-41d4-a716-446655440000-ui-design.png",
  "image_url": "https://localhost:9000/ui-screenshots/screenshots/550e8400-e29b-41d4-a716-446655440000-ui-design.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=...",
  "ground_truth": {
    "elements": [
      {"type": "button", "text": "Submit", "position": {"x": 100, "y": 200}},
      {"type": "input", "placeholder": "Enter text", "position": {"x": 100, "y": 150}}
    ]
  },
  "predicted_labels": null,
  "evaluation_scores": null,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z",
  "file_info": {
    "size": 245760,
    "content_type": "image/png"
  }
}
```

### Get All Images
```
GET /api/v1/images/
```

### Get Image by ID
```
GET /api/v1/images/{id}
```

### Update Image Predictions
```
PUT /api/v1/images/{id}
Content-Type: application/json

{
  "predicted_labels": {
    "model1": {"label": "button", "confidence": 0.95},
    "model2": {"label": "input", "confidence": 0.87}
  },
  "evaluation_scores": {
    "accuracy": 0.92,
    "precision": 0.89,
    "recall": 0.94
  }
}
```

### Update Image Ground Truth
```
PUT /api/v1/images/{id}/ground-truth
Content-Type: application/json

{
  "ground_truth": {
    "elements": [
      {"type": "button", "text": "Submit", "position": {"x": 100, "y": 200}},
      {"type": "input", "placeholder": "Enter text", "position": {"x": 100, "y": 150}}
    ]
  }
}
```

### Delete Image
```