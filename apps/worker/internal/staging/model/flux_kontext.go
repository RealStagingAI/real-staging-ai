package model

import (
	"context"
	"fmt"

	"github.com/replicate/replicate-go"
)

// FluxKontextInputBuilder builds input parameters for the Flux Kontext Max model.
type FluxKontextInputBuilder struct{}

// Ensure FluxKontextInputBuilder implements ModelInputBuilder.
var _ ModelInputBuilder = (*FluxKontextInputBuilder)(nil)

// NewFluxKontextInputBuilder creates a new FluxKontextInputBuilder.
func NewFluxKontextInputBuilder() *FluxKontextInputBuilder {
	return &FluxKontextInputBuilder{}
}

// BuildInput creates the input parameters for the Flux Kontext model.
func (b *FluxKontextInputBuilder) BuildInput(
	ctx context.Context, req *ModelInputRequest,
) (replicate.PredictionInput, error) {
	if err := b.Validate(req); err != nil {
		return nil, err
	}

	// Use provided config or fall back to defaults
	config := req.Config
	if config == nil {
		config = (&FluxKontextConfig{}).GetDefaults()
	}

	// Type assert to FluxKontextConfig
	fluxConfig, ok := config.(*FluxKontextConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type for Flux Kontext model")
	}

	// Build input from config
	input := replicate.PredictionInput{
		"prompt":            req.Prompt,
		"input_image":       req.ImageDataURL,
		"aspect_ratio":      fluxConfig.AspectRatio,
		"output_format":     fluxConfig.OutputFormat,
		"safety_tolerance":  fluxConfig.SafetyTolerance,
		"prompt_upsampling": fluxConfig.PromptUpsampling,
		"num_outputs":       fluxConfig.NumOutputs,
		"output_quality":    fluxConfig.OutputQuality,
	}

	// Seed from config takes precedence over request seed
	if fluxConfig.Seed != nil {
		input["seed"] = *fluxConfig.Seed
	} else if req.Seed != nil {
		input["seed"] = *req.Seed
	}

	return input, nil
}

// Validate checks if the request is valid for the Flux Kontext Max model.
func (b *FluxKontextInputBuilder) Validate(req *ModelInputRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if req.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}
	// Note: input_image is optional for Flux Kontext Max
	return nil
}
