package model

import (
	"context"
	"fmt"

	"github.com/replicate/replicate-go"
)

// SeedreamInputBuilder builds input parameters for ByteDance Seedream models.
// This builder supports all Seedream versions (Seedream-3, Seedream-4, and future versions)
// as they share a common API structure.
type SeedreamInputBuilder struct{}

// Ensure SeedreamInputBuilder implements ModelInputBuilder.
var _ ModelInputBuilder = (*SeedreamInputBuilder)(nil)

// NewSeedreamInputBuilder creates a new SeedreamInputBuilder.
func NewSeedreamInputBuilder() *SeedreamInputBuilder {
	return &SeedreamInputBuilder{}
}

// BuildInput creates the input parameters for Seedream models.
func (b *SeedreamInputBuilder) BuildInput(
	ctx context.Context, req *ModelInputRequest,
) (replicate.PredictionInput, error) {
	if err := b.Validate(req); err != nil {
		return nil, err
	}

	input := replicate.PredictionInput{
		"prompt":         req.Prompt,
		"size":           "big",    // Options: "small", "regular", "big"
		"aspect_ratio":   "custom", // Use custom to preserve input image dimensions
		"enhance_prompt": true,     // Enable prompt enhancement for better quality
		"max_images":     1,        // Single image output
	}

	// Add input image if provided
	if req.ImageDataURL != "" {
		input["image_input"] = []string{req.ImageDataURL}
	}

	// Add seed if provided
	if req.Seed != nil {
		input["seed"] = *req.Seed
	}

	return input, nil
}

// Validate checks if the request is valid for Seedream models.
func (b *SeedreamInputBuilder) Validate(req *ModelInputRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if req.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}
	// Note: ImageDataURL is optional - model supports both text-to-image and image-to-image
	return nil
}
