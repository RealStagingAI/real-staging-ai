package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/replicate/replicate-go"
)

// GPTImageInputBuilder builds input parameters for OpenAI GPT Image models.
type GPTImageInputBuilder struct{}

// Ensure GPTImageInputBuilder implements ModelInputBuilder.
var _ ModelInputBuilder = (*GPTImageInputBuilder)(nil)

// NewGPTImageInputBuilder creates a new GPTImageInputBuilder.
func NewGPTImageInputBuilder() *GPTImageInputBuilder {
	return &GPTImageInputBuilder{}
}

// BuildInput creates the input parameters for GPT Image models.
func (b *GPTImageInputBuilder) BuildInput(
	ctx context.Context, req *ModelInputRequest,
) (replicate.PredictionInput, error) {
	if err := b.Validate(req); err != nil {
		return nil, err
	}

	config := req.Config
	if config == nil {
		config = (&GPTImageConfig{}).GetDefaults()
	}

	gptConfig, ok := config.(*GPTImageConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type for GPT Image model")
	}

	if strings.TrimSpace(gptConfig.OpenAIAPIKey) == "" {
		return nil, fmt.Errorf("openai_api_key is required")
	}

	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		prompt = strings.TrimSpace(gptConfig.Prompt)
	}
	if prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	input := replicate.PredictionInput{
		"openai_api_key":     gptConfig.OpenAIAPIKey,
		"prompt":             prompt,
		"quality":            gptConfig.Quality,
		"aspect_ratio":       gptConfig.AspectRatio,
		"input_fidelity":     gptConfig.InputFidelity,
		"number_of_images":   gptConfig.NumberOfImages,
		"background":         gptConfig.Background,
		"output_compression": gptConfig.OutputCompression,
		"output_format":      gptConfig.OutputFormat,
		"moderation":         gptConfig.Moderation,
	}

	// Add input image if provided (from the request, not config)
	if trimmed := strings.TrimSpace(req.ImageDataURL); trimmed != "" {
		input["input_images"] = []string{trimmed}
	}

	if gptConfig.UserID != nil {
		if userID := strings.TrimSpace(*gptConfig.UserID); userID != "" {
			input["user_id"] = userID
		}
	}

	return input, nil
}

// Validate checks if the request is valid for GPT Image models.
func (b *GPTImageInputBuilder) Validate(req *ModelInputRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	// Image input is optional; model supports both text-to-image and image-to-image flows.
	return nil
}
