package originalimage

import (
	"context"
	"time"

	"github.com/real-staging-ai/api/internal/storage/queries"
)

//go:generate go run github.com/matryer/moq@v0.5.3 -out repository_mock.go . Repository

// Repository defines the interface for original image data access operations.
type Repository interface {
	// CreateOriginalImage creates a new original image record.
	CreateOriginalImage(
		ctx context.Context,
		contentHash, s3Key string,
		fileSize int64,
		mimeType string,
		width, height *int,
	) (*queries.OriginalImage, error)

	// GetOriginalImageByID retrieves an original image by its ID.
	GetOriginalImageByID(ctx context.Context, id string) (*queries.OriginalImage, error)

	// GetOriginalImageByHash retrieves an original image by its content hash.
	GetOriginalImageByHash(ctx context.Context, contentHash string) (*queries.OriginalImage, error)

	// IncrementReferenceCount increments the reference count for an original image.
	IncrementReferenceCount(ctx context.Context, id string) error

	// DecrementReferenceCount decrements the reference count for an original image.
	// If the reference count reaches 0, the image becomes eligible for cleanup.
	DecrementReferenceCount(ctx context.Context, id string) error

	// ListOrphanedOriginalImages lists original images with zero references older than the given duration.
	ListOrphanedOriginalImages(
		ctx context.Context,
		olderThan time.Duration,
		limit int,
	) ([]*queries.OriginalImage, error)

	// DeleteOriginalImage permanently deletes an original image record.
	DeleteOriginalImage(ctx context.Context, id string) error

	// GetOriginalImageStats retrieves statistics about original images.
	GetOriginalImageStats(ctx context.Context) (*OriginalImageStats, error)
}

// OriginalImageStats contains statistics about original images.
type OriginalImageStats struct {
	TotalCount    int64
	TotalSize     int64
	OrphanedCount int64
	OrphanedSize  int64
	AvgReferences float64
}
