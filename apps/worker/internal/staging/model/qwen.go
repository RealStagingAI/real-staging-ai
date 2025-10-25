package model

import (
	"context"
	"fmt"

	"github.com/replicate/replicate-go"
)

// QwenInputBuilder builds input parameters for the Qwen Image Edit model.
type QwenInputBuilder struct{}

// Ensure QwenInputBuilder implements ModelInputBuilder.
var _ ModelInputBuilder = (*QwenInputBuilder)(nil)

// NewQwenInputBuilder creates a new QwenInputBuilder.
func NewQwenInputBuilder() *QwenInputBuilder {
	return &QwenInputBuilder{}
}

// BuildInput creates the input parameters for the Qwen Image Edit model.
func (b *QwenInputBuilder) BuildInput(ctx context.Context, req *ModelInputRequest) (replicate.PredictionInput, error) {
	if err := b.Validate(req); err != nil {
		return nil, err
	}

	// Use provided config or fall back to defaults
	config := req.Config
	if config == nil {
		config = (&QwenConfig{}).GetDefaults()
	}

	// Type assert to QwenConfig
	qwenConfig, ok := config.(*QwenConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type for Qwen model")
	}

	// Build input from config
	input := replicate.PredictionInput{
		"image":          req.ImageDataURL,
		"prompt":         req.Prompt,
		"go_fast":        qwenConfig.GoFast,
		"aspect_ratio":   qwenConfig.AspectRatio,
		"output_format":  qwenConfig.OutputFormat,
		"output_quality": qwenConfig.OutputQuality,
	}

	// Seed from config takes precedence over request seed
	if qwenConfig.Seed != nil {
		input["seed"] = *qwenConfig.Seed
	} else if req.Seed != nil {
		input["seed"] = *req.Seed
	}

	return input, nil
}

// Validate checks if the request is valid for the Qwen Image Edit model.
func (b *QwenInputBuilder) Validate(req *ModelInputRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if req.ImageDataURL == "" {
		return fmt.Errorf("image data URL is required")
	}
	if req.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}
	return nil
}
