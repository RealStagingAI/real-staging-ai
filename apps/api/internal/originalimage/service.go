package originalimage

import (
	"context"
	"fmt"
	"time"

	"github.com/real-staging-ai/api/internal/storage"
)

//go:generate go run github.com/matryer/moq@v0.5.3 -out service_mock.go . Service

// Service defines the interface for original image business logic operations.
type Service interface {
	// DecrementReferenceAndCleanup decrements the reference count for an original image
	// and deletes it from both database and S3 if the count reaches zero.
	// Returns true if the original was deleted, false otherwise.
	DecrementReferenceAndCleanup(ctx context.Context, originalImageID string) (bool, error)

	// CleanupOrphanedOriginals finds and deletes original images with zero references
	// that are older than the specified duration. Returns the number of images cleaned up.
	CleanupOrphanedOriginals(ctx context.Context, olderThan time.Duration, limit int) (int, error)

	// GetStats retrieves statistics about original images.
	GetStats(ctx context.Context) (*OriginalImageStats, error)
}

// DefaultService implements the Service interface.
type DefaultService struct {
	repo      Repository
	s3Service storage.S3Service
}

// NewDefaultService creates a new DefaultService instance.
func NewDefaultService(repo Repository, s3Service storage.S3Service) *DefaultService {
	return &DefaultService{
		repo:      repo,
		s3Service: s3Service,
	}
}

// DecrementReferenceAndCleanup decrements the reference count and cleans up if needed.
func (s *DefaultService) DecrementReferenceAndCleanup(ctx context.Context, originalImageID string) (bool, error) {
	if originalImageID == "" {
		return false, fmt.Errorf("original image ID cannot be empty")
	}

	// Get the original image before decrementing to check reference count
	original, err := s.repo.GetOriginalImageByID(ctx, originalImageID)
	if err != nil {
		return false, fmt.Errorf("failed to get original image: %w", err)
	}

	// Decrement the reference count
	if err := s.repo.DecrementReferenceCount(ctx, originalImageID); err != nil {
		return false, fmt.Errorf("failed to decrement reference count: %w", err)
	}

	// If this was the last reference, clean up the original
	if original.ReferenceCount <= 1 {
		// Delete from S3
		// Ignore S3 deletion errors - database cleanup is more important
		// The orphaned file can be cleaned up later by background job
		_ = s.s3Service.DeleteFile(ctx, original.S3Key)

		// Delete from database
		if err := s.repo.DeleteOriginalImage(ctx, originalImageID); err != nil {
			return false, fmt.Errorf("failed to delete original image: %w", err)
		}

		return true, nil
	}

	return false, nil
}

// CleanupOrphanedOriginals finds and deletes orphaned original images.
func (s *DefaultService) CleanupOrphanedOriginals(
	ctx context.Context, olderThan time.Duration, limit int,
) (int, error) {
	orphaned, err := s.repo.ListOrphanedOriginalImages(ctx, olderThan, limit)
	if err != nil {
		return 0, fmt.Errorf("failed to list orphaned originals: %w", err)
	}

	deletedCount := 0
	for _, original := range orphaned {
		// Delete from S3
		if deleteErr := s.s3Service.DeleteFile(ctx, original.S3Key); deleteErr != nil {
			// Log but continue with other deletions
			// TODO: Add logging when logger is available
			continue
		}

		// Delete from database
		originalID := formatUUID(original.ID.Bytes)
		if err := s.repo.DeleteOriginalImage(ctx, originalID); err != nil {
			// Log but continue with other deletions
			// TODO: Add logging when logger is available
			continue
		}

		deletedCount++
	}

	return deletedCount, nil
}

// GetStats retrieves statistics about original images.
func (s *DefaultService) GetStats(ctx context.Context) (*OriginalImageStats, error) {
	stats, err := s.repo.GetOriginalImageStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return stats, nil
}

// formatUUID converts [16]byte UUID to string format.
func formatUUID(b [16]byte) string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
