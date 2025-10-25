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

	// Use provided config or fall back to defaults
	config := req.Config
	if config == nil {
		config = (&SeedreamConfig{}).GetDefaults()
	}

	// Type assert to SeedreamConfig
	seedreamConfig, ok := config.(*SeedreamConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type for Seedream model")
	}

	// Build input from config
	input := replicate.PredictionInput{
		"prompt":              req.Prompt,
		"aspect_ratio":        seedreamConfig.AspectRatio,
		"num_inference_steps": seedreamConfig.NumInferenceSteps,
		"guidance_scale":      seedreamConfig.GuidanceScale,
		"output_quality":      seedreamConfig.OutputQuality,
	}

	// Add input image if provided
	if req.ImageDataURL != "" {
		input["image_input"] = []string{req.ImageDataURL}
	}

	// Seed from config takes precedence over request seed
	if seedreamConfig.Seed != nil {
		input["seed"] = *seedreamConfig.Seed
	} else if req.Seed != nil {
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
