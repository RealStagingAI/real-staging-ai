package originalimage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/real-staging-ai/api/internal/storage"
	"github.com/real-staging-ai/api/internal/storage/queries"
)

// DefaultRepository implements the Repository interface.
type DefaultRepository struct {
	db storage.Database
}

// NewDefaultRepository creates a new DefaultRepository instance.
func NewDefaultRepository(db storage.Database) *DefaultRepository {
	return &DefaultRepository{db: db}
}

// CreateOriginalImage creates a new original image record in the database.
func (r *DefaultRepository) CreateOriginalImage(
	ctx context.Context,
	contentHash, s3Key string,
	fileSize int64,
	mimeType string,
	width, height *int,
) (*queries.OriginalImage, error) {
	q := queries.New(r.db)

	var widthInt, heightInt pgtype.Int4
	if width != nil {
		widthInt = pgtype.Int4{Int32: int32(*width), Valid: true} // #nosec G115 -- image dimensions are small
	}
	if height != nil {
		heightInt = pgtype.Int4{Int32: int32(*height), Valid: true} // #nosec G115 -- image dimensions are small
	}

	result, err := q.CreateOriginalImage(ctx, queries.CreateOriginalImageParams{
		ContentHash: contentHash,
		S3Key:       s3Key,
		FileSize:    fileSize,
		MimeType:    mimeType,
		Width:       widthInt,
		Height:      heightInt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create original image: %w", err)
	}

	return result, nil
}

// GetOriginalImageByID retrieves an original image by its ID.
func (r *DefaultRepository) GetOriginalImageByID(ctx context.Context, id string) (*queries.OriginalImage, error) {
	q := queries.New(r.db)

	imageUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid original image ID: %w", err)
	}

	result, err := q.GetOriginalImageByID(ctx, pgtype.UUID{Bytes: imageUUID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get original image: %w", err)
	}

	return result, nil
}

// GetOriginalImageByHash retrieves an original image by its content hash.
func (r *DefaultRepository) GetOriginalImageByHash(
	ctx context.Context, contentHash string,
) (*queries.OriginalImage, error) {
	q := queries.New(r.db)

	result, err := q.GetOriginalImageByHash(ctx, contentHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get original image by hash: %w", err)
	}

	return result, nil
}

// IncrementReferenceCount increments the reference count for an original image.
func (r *DefaultRepository) IncrementReferenceCount(ctx context.Context, id string) error {
	q := queries.New(r.db)

	imageUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid original image ID: %w", err)
	}

	if err := q.IncrementReferenceCount(ctx, pgtype.UUID{Bytes: imageUUID, Valid: true}); err != nil {
		return fmt.Errorf("failed to increment reference count: %w", err)
	}

	return nil
}

// DecrementReferenceCount decrements the reference count for an original image.
func (r *DefaultRepository) DecrementReferenceCount(ctx context.Context, id string) error {
	q := queries.New(r.db)

	imageUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid original image ID: %w", err)
	}

	if err := q.DecrementReferenceCount(ctx, pgtype.UUID{Bytes: imageUUID, Valid: true}); err != nil {
		return fmt.Errorf("failed to decrement reference count: %w", err)
	}

	return nil
}

// ListOrphanedOriginalImages lists original images with zero references older than the given duration.
func (r *DefaultRepository) ListOrphanedOriginalImages(
	ctx context.Context,
	olderThan time.Duration,
	limit int,
) ([]*queries.OriginalImage, error) {
	q := queries.New(r.db)

	interval := pgtype.Interval{
		Microseconds: olderThan.Microseconds(),
		Valid:        true,
	}

	results, err := q.ListOrphanedOriginalImages(ctx, queries.ListOrphanedOriginalImagesParams{
		Column1: interval,
		Limit:   int32(limit), // #nosec G115 -- limit is controlled by caller
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list orphaned original images: %w", err)
	}

	return results, nil
}

// DeleteOriginalImage permanently deletes an original image record.
func (r *DefaultRepository) DeleteOriginalImage(ctx context.Context, id string) error {
	q := queries.New(r.db)

	imageUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid original image ID: %w", err)
	}

	if err := q.DeleteOriginalImage(ctx, pgtype.UUID{Bytes: imageUUID, Valid: true}); err != nil {
		return fmt.Errorf("failed to delete original image: %w", err)
	}

	return nil
}

// GetOriginalImageStats retrieves statistics about original images.
func (r *DefaultRepository) GetOriginalImageStats(ctx context.Context) (*OriginalImageStats, error) {
	q := queries.New(r.db)

	result, err := q.GetOriginalImageStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get original image stats: %w", err)
	}

	stats := &OriginalImageStats{
		TotalCount:    result.TotalCount,
		TotalSize:     result.TotalSize,
		OrphanedCount: result.OrphanedCount,
		OrphanedSize:  result.OrphanedSize,
		AvgReferences: result.AvgReferences,
	}

	return stats, nil
}
