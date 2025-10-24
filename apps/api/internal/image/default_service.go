package image

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/real-staging-ai/api/internal/config"
	"github.com/real-staging-ai/api/internal/job"
	"github.com/real-staging-ai/api/internal/logging"
	"github.com/real-staging-ai/api/internal/queue"
	"github.com/real-staging-ai/api/internal/storage/queries"
)

var jsonMarshal = json.Marshal

// OriginalImageService defines the interface for original image operations.
// This is a minimal interface to avoid circular dependencies.
type OriginalImageService interface {
	DecrementReferenceAndCleanup(ctx context.Context, originalImageID string) (bool, error)
}

// DefaultService handles business logic for image operations.
type DefaultService struct {
	imageRepo            Repository
	jobRepo              job.Repository
	enqueuer             queue.Enqueuer
	originalImageService OriginalImageService
}

// NewDefaultService creates a new DefaultService instance.
func NewDefaultService(
	cfg *config.Config,
	imageRepo Repository,
	jobRepo job.Repository,
	originalImageService OriginalImageService,
) *DefaultService {
	// Best-effort build an enqueuer from env or config; fall back to Noop if not configured.
	var enq queue.Enqueuer
	if e, err := queue.NewAsynqEnqueuerFromEnv(cfg); err == nil {
		enq = e
	} else {
		enq = queue.NoopEnqueuer{}
	}
	return &DefaultService{
		imageRepo:            imageRepo,
		jobRepo:              jobRepo,
		enqueuer:             enq,
		originalImageService: originalImageService,
	}
}

// CreateImage creates a new image and queues it for processing.
func (s *DefaultService) CreateImage(ctx context.Context, req *CreateImageRequest) (*Image, error) {
	log := logging.NewDefaultLogger()
	if req == nil {
		err := fmt.Errorf("request cannot be nil")
		log.Error(ctx, "create image: invalid request", "error", err)
		return nil, err
	}

	// Create the image in the database
	dbImage, err := s.imageRepo.CreateImage(
		ctx,
		req.ProjectID.String(),
		req.OriginalURL,
		req.RoomType,
		req.Style,
		req.Seed,
		req.Prompt,
	)
	if err != nil {
		log.Error(ctx, "create image: repo failure",
			"project_id", req.ProjectID.String(),
			"original_url", req.OriginalURL,
			"error", err)
		return nil, fmt.Errorf("failed to create image: %w", err)
	}

	// Convert database image to domain image
	domainImage := s.convertToImage(dbImage)

	// Create job payload
	payload := JobPayload{
		ImageID:     domainImage.ID,
		OriginalURL: domainImage.OriginalURL,
		RoomType:    domainImage.RoomType,
		Style:       domainImage.Style,
		Seed:        domainImage.Seed,
		Prompt:      domainImage.Prompt,
	}
	payloadJSON, err := jsonMarshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal job payload: %w", err)
	}

	// Create a job for processing the image (persist metadata)
	_, err = s.jobRepo.CreateJob(ctx, domainImage.ID.String(), "stage:run", payloadJSON)
	if err != nil {
		log.Error(ctx, "create image: job create failed", "image_id", domainImage.ID.String(), "error", err)
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	// Enqueue processing task to the queue
	log.Info(ctx, "enqueue stage:run", "image_id", domainImage.ID.String())
	if _, err := s.enqueuer.EnqueueStageRun(ctx, queue.StageRunPayload{
		ImageID:     domainImage.ID.String(),
		OriginalURL: domainImage.OriginalURL,
		RoomType:    domainImage.RoomType,
		Style:       domainImage.Style,
		Seed:        domainImage.Seed,
		Prompt:      domainImage.Prompt,
	}, nil); err != nil {
		log.Error(ctx, "enqueue stage:run failed", "image_id", domainImage.ID.String(), "error", err)
		return nil, fmt.Errorf("failed to enqueue stage:run: %w", err)
	}
	log.Info(ctx, "image enqueued", "image_id", domainImage.ID.String())

	return domainImage, nil
}

// BatchCreateImages creates multiple images in a single transaction.
func (s *DefaultService) BatchCreateImages(
	ctx context.Context, reqs []CreateImageRequest,
) (*BatchCreateImagesResponse, error) {
	log := logging.NewDefaultLogger()

	response := &BatchCreateImagesResponse{
		Images: []*Image{},
		Errors: []BatchImageError{},
	}

	// Process each image request
	for i, req := range reqs {
		img, err := s.CreateImage(ctx, &req)
		if err != nil {
			log.Error(ctx, "batch create: failed to create image",
				"index", i,
				"project_id", req.ProjectID.String(),
				"error", err)
			return nil, fmt.Errorf("failed to create image at index %d: %w", i, err)
		} else {
			response.Images = append(response.Images, img)
		}
	}

	log.Info(ctx, "batch create completed",
		"total", len(reqs),
		"success", len(response.Images),
		"failed", len(response.Errors))
	return response, nil
}

// GetImageByID retrieves a specific image by its ID.
func (s *DefaultService) GetImageByID(ctx context.Context, imageID string) (*Image, error) {
	if imageID == "" {
		return nil, fmt.Errorf("image ID cannot be empty")
	}

	dbImage, err := s.imageRepo.GetImageByID(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	return s.convertToImage(dbImage), nil
}

// GetImagesByProjectID retrieves all images for a specific project.
func (s *DefaultService) GetImagesByProjectID(ctx context.Context, projectID string) ([]*Image, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID cannot be empty")
	}

	dbImages, err := s.imageRepo.GetImagesByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get images: %w", err)
	}

	images := make([]*Image, len(dbImages))
	for i, dbImage := range dbImages {
		images[i] = s.convertToImage(dbImage)
	}

	return images, nil
}

// GetGroupedProjectImages retrieves images grouped by original_image_id.
func (s *DefaultService) GetGroupedProjectImages(
	ctx context.Context, projectID string,
) (*GroupedProjectImagesResponse, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID cannot be empty")
	}

	dbImages, err := s.imageRepo.GetImagesByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get images: %w", err)
	}

	// Group images by original_image_id (or by original_url if no original_image_id)
	groupMap := make(map[string]*GroupedImage)
	for _, dbImage := range dbImages {
		// Use original_image_id as key if available, else use original_url
		var groupKey string
		switch {
		case dbImage.OriginalImageID.Valid:
			groupKey = dbImage.OriginalImageID.String()
		case dbImage.OriginalUrl.Valid:
			groupKey = dbImage.OriginalUrl.String
		default:
			// Skip images without original_url (shouldn't happen)
			continue
		}

		if _, exists := groupMap[groupKey]; !exists {
			// Create new group
			var originalImageID *string
			if dbImage.OriginalImageID.Valid {
				oidStr := dbImage.OriginalImageID.String()
				originalImageID = &oidStr
			}

			originalURL := ""
			if dbImage.OriginalUrl.Valid {
				originalURL = dbImage.OriginalUrl.String
			}

			group := &GroupedImage{
				OriginalImageID: originalImageID,
				OriginalURL:     originalURL,
				Variants:        []*ImageVariant{},
			}

			if dbImage.RoomType.Valid {
				group.RoomType = &dbImage.RoomType.String
			}
			if dbImage.Seed.Valid {
				group.Seed = &dbImage.Seed.Int64
			}
			if dbImage.Prompt.Valid {
				group.Prompt = &dbImage.Prompt.String
			}

			groupMap[groupKey] = group
		}

		// Add variant to group
		variant := &ImageVariant{
			ID:        dbImage.ID.Bytes,
			Status:    Status(dbImage.Status),
			CreatedAt: dbImage.CreatedAt.Time,
			UpdatedAt: dbImage.UpdatedAt.Time,
		}

		if dbImage.Style.Valid {
			variant.Style = &dbImage.Style.String
		}
		if dbImage.StagedUrl.Valid {
			variant.StagedURL = &dbImage.StagedUrl.String
		}
		if dbImage.Error.Valid {
			variant.Error = &dbImage.Error.String
		}
		// TODO: Convert pgtype.Numeric to float64 for CostUSD when needed
		// if dbImage.CostUsd.Valid { variant.CostUSD = ... }
		if dbImage.ProcessingTimeMs.Valid {
			processingTimeMs := int(dbImage.ProcessingTimeMs.Int32)
			variant.ProcessingTimeMs = &processingTimeMs
		}
		if dbImage.ModelUsed.Valid {
			variant.ModelUsed = &dbImage.ModelUsed.String
		}
		if dbImage.ReplicatePredictionID.Valid {
			variant.ReplicatePredictionID = &dbImage.ReplicatePredictionID.String
		}

		groupMap[groupKey].Variants = append(groupMap[groupKey].Variants, variant)
	}

	// Convert map to slice
	images := make([]*GroupedImage, 0, len(groupMap))
	for _, group := range groupMap {
		images = append(images, group)
	}

	return &GroupedProjectImagesResponse{Images: images}, nil
}

// UpdateImageStatus updates an image's processing status.
func (s *DefaultService) UpdateImageStatus(ctx context.Context, imageID string, status Status) (*Image, error) {
	if imageID == "" {
		return nil, fmt.Errorf("image ID cannot be empty")
	}

	dbImage, err := s.imageRepo.UpdateImageStatus(ctx, imageID, status.String())
	if err != nil {
		return nil, fmt.Errorf("failed to update image status: %w", err)
	}

	return s.convertToImage(dbImage), nil
}

// UpdateImageWithStagedURL updates an image with the staged URL and marks it as ready.
func (s *DefaultService) UpdateImageWithStagedURL(
	ctx context.Context, imageID string, stagedURL string,
) (*Image, error) {
	if imageID == "" {
		return nil, fmt.Errorf("image ID cannot be empty")
	}
	if stagedURL == "" {
		return nil, fmt.Errorf("staged URL cannot be empty")
	}

	dbImage, err := s.imageRepo.UpdateImageWithStagedURL(ctx, imageID, stagedURL, StatusReady.String())
	if err != nil {
		return nil, fmt.Errorf("failed to update image with staged URL: %w", err)
	}

	return s.convertToImage(dbImage), nil
}

// UpdateImageWithError updates an image with an error status and message.
func (s *DefaultService) UpdateImageWithError(ctx context.Context, imageID string, errorMsg string) (*Image, error) {
	if imageID == "" {
		return nil, fmt.Errorf("image ID cannot be empty")
	}
	if errorMsg == "" {
		return nil, fmt.Errorf("error message cannot be empty")
	}

	dbImage, err := s.imageRepo.UpdateImageWithError(ctx, imageID, errorMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to update image with error: %w", err)
	}

	return s.convertToImage(dbImage), nil
}

// DeleteImage deletes an image from the database and decrements the original image reference.
// If this is the last reference to the original image, the original is also deleted from S3 and database.
func (s *DefaultService) DeleteImage(ctx context.Context, imageID string) error {
	log := logging.Default()
	if imageID == "" {
		return fmt.Errorf("image ID cannot be empty")
	}

	// Get the original_image_id before soft-deleting the image
	originalImageID, err := s.imageRepo.GetOriginalImageID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("failed to get original image ID: %w", err)
	}

	// Soft delete the image (marks as deleted but keeps for billing/usage tracking)
	err = s.imageRepo.DeleteImage(ctx, imageID)
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	// If the image has an associated original, decrement its reference count
	// and clean up if this was the last reference
	if originalImageID != "" && s.originalImageService != nil {
		deleted, err := s.originalImageService.DecrementReferenceAndCleanup(ctx, originalImageID)
		if err != nil {
			// Log the error but don't fail the deletion
			// The orphaned original can be cleaned up later by a background job
			log.Warn(ctx, "failed to decrement original image reference", "original_id", originalImageID, "error", err)
		} else if deleted {
			log.Info(ctx, "deleted unreferenced original image", "original_id", originalImageID)
		}
	}

	return nil
}

// convertToImage converts a database image to a domain image.
func (s *DefaultService) convertToImage(dbImage *queries.Image) *Image {
	// Extract OriginalURL - default to empty string if null (for migration compatibility)
	originalURL := ""
	if dbImage.OriginalUrl.Valid {
		originalURL = dbImage.OriginalUrl.String
	}

	image := &Image{
		ID:          dbImage.ID.Bytes,
		ProjectID:   dbImage.ProjectID.Bytes,
		OriginalURL: originalURL,
		Status:      Status(dbImage.Status),
		CreatedAt:   dbImage.CreatedAt.Time,
		UpdatedAt:   dbImage.UpdatedAt.Time,
	}

	if dbImage.StagedUrl.Valid {
		image.StagedURL = &dbImage.StagedUrl.String
	}

	if dbImage.RoomType.Valid {
		image.RoomType = &dbImage.RoomType.String
	}

	if dbImage.Style.Valid {
		image.Style = &dbImage.Style.String
	}

	if dbImage.Seed.Valid {
		image.Seed = &dbImage.Seed.Int64
	}

	if dbImage.Prompt.Valid {
		image.Prompt = &dbImage.Prompt.String
	}

	if dbImage.Error.Valid {
		image.Error = &dbImage.Error.String
	}

	return image
}

// GetProjectCostSummary retrieves cost summary for a project.
func (s *DefaultService) GetProjectCostSummary(ctx context.Context, projectID string) (*ProjectCostSummary, error) {
	return s.imageRepo.GetProjectCostSummary(ctx, projectID)
}
