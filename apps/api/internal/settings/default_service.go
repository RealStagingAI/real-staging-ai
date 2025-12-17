package settings

import (
	"context"
	"encoding/json"
	"fmt"
)

// DefaultService implements Service.
type DefaultService struct {
	repo Repository
}

// Ensure DefaultService implements Service.
var _ Service = (*DefaultService)(nil)

// NewDefaultService creates a new DefaultService.
func NewDefaultService(repo Repository) *DefaultService {
	return &DefaultService{repo: repo}
}

// GetActiveModel retrieves the currently active AI model ID.
func (s *DefaultService) GetActiveModel(ctx context.Context) (string, error) {
	setting, err := s.repo.GetByKey(ctx, "active_model")
	if err != nil {
		return "", fmt.Errorf("failed to get active model: %w", err)
	}
	return setting.Value, nil
}

// UpdateActiveModel updates the active AI model.
func (s *DefaultService) UpdateActiveModel(ctx context.Context, modelID, userID string) error {
	// Validate model ID against available models
	models, err := s.ListAvailableModels(ctx)
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	valid := false
	for _, model := range models {
		if model.ID == modelID {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid model ID: %s", modelID)
	}

	return s.repo.Update(ctx, "active_model", modelID, userID)
}

// ListAvailableModels returns all available AI models.
// This is hardcoded for now but could be made dynamic in the future.
func (s *DefaultService) ListAvailableModels(ctx context.Context) ([]ModelInfo, error) {
	// Get the current active model
	activeModelID, _ := s.GetActiveModel(ctx)

	models := []ModelInfo{
		{
			ID:          "qwen/qwen-image-edit",
			Name:        "Qwen Image Edit",
			Description: "Fast image editing model optimized for virtual staging. Requires input image.",
			Version:     "v1",
			IsActive:    activeModelID == "qwen/qwen-image-edit",
		},
		{
			ID:   "black-forest-labs/flux-kontext-max",
			Name: "Flux Kontext Max",
			Description: "High-quality image generation and editing with advanced context understanding. " +
				"Supports both text-to-image and image-to-image.",
			Version:  "v1",
			IsActive: activeModelID == "black-forest-labs/flux-kontext-max",
		},
		{
			ID:   "black-forest-labs/flux-kontext-pro",
			Name: "Flux Kontext Pro",
			Description: "State-of-the-art text-based image editing with high-quality outputs and excellent prompt following. " +
				"Professional-grade editing capabilities.",
			Version:  "v1",
			IsActive: activeModelID == "black-forest-labs/flux-kontext-pro",
		},
		{
			ID:   "bytedance/seedream-3",
			Name: "Seedream 3",
			Description: "Unified text-to-image generation and precise editing. " +
				"Supports both workflows with natural language commands.",
			Version:  "v1",
			IsActive: activeModelID == "bytedance/seedream-3",
		},
		{
			ID:   "bytedance/seedream-4",
			Name: "Seedream 4",
			Description: "Latest Seedream model with support for up to 4K resolution. " +
				"High-quality text-to-image and image editing.",
			Version:  "v1",
			IsActive: activeModelID == "bytedance/seedream-4",
		},
		{
			ID:   "openai/gpt-image-1",
			Name: "GPT Image 1",
			Description: "A multimodal image generation model that creates high-quality images. " +
				"You need to bring your own verified OpenAI key to use this model. " +
				"Your OpenAI account will be charged for usage.",
			Version:  "v1",
			IsActive: activeModelID == "openai/gpt-image-1",
		},
		{
			ID:   "openai/gpt-image-1.5",
			Name: "GPT Image 1.5",
			Description: "A multimodal image generation model that creates high-quality images. " +
				"You need to bring your own verified OpenAI key to use this model. " +
				"Your OpenAI account will be charged for usage.",
			Version:  "v1",
			IsActive: activeModelID == "openai/gpt-image-1.5",
		},
	}

	return models, nil
}

// GetSetting retrieves a setting by key.
func (s *DefaultService) GetSetting(ctx context.Context, key string) (*Setting, error) {
	return s.repo.GetByKey(ctx, key)
}

// UpdateSetting updates a setting value.
func (s *DefaultService) UpdateSetting(ctx context.Context, key, value, userID string) error {
	return s.repo.Update(ctx, key, value, userID)
}

// ListSettings retrieves all settings.
func (s *DefaultService) ListSettings(ctx context.Context) ([]Setting, error) {
	return s.repo.List(ctx)
}

// GetModelConfig retrieves the configuration for a specific model.
func (s *DefaultService) GetModelConfig(ctx context.Context, modelID string) (*ModelConfig, error) {
	configJSON, err := s.repo.GetModelConfig(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get model config: %w", err)
	}

	// Parse JSON into map
	var configMap map[string]interface{}
	if err := json.Unmarshal(configJSON, &configMap); err != nil {
		return nil, fmt.Errorf("failed to parse model config: %w", err)
	}

	return &ModelConfig{
		ModelID: modelID,
		Config:  configMap,
	}, nil
}

// UpdateModelConfig updates the configuration for a specific model.
func (s *DefaultService) UpdateModelConfig(
	ctx context.Context, modelID string, config map[string]interface{}, userID string,
) error {
	// Validate model exists
	models, err := s.ListAvailableModels(ctx)
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	valid := false
	for _, model := range models {
		if model.ID == modelID {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid model ID: %s", modelID)
	}

	// Marshal config to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Update in database
	return s.repo.UpdateModelConfig(ctx, modelID, configJSON, userID)
}

// GetModelConfigSchema returns the schema for a model's configuration.
func (s *DefaultService) GetModelConfigSchema(ctx context.Context, modelID string) (*ModelConfigSchema, error) {
	// Return schema based on model ID
	switch modelID {
	case "qwen/qwen-image-edit":
		return getQwenSchema(), nil
	case "black-forest-labs/flux-kontext-max", "black-forest-labs/flux-kontext-pro":
		return getFluxKontextSchema(modelID), nil
	case "bytedance/seedream-3", "bytedance/seedream-4":
		return getSeedreamSchema(modelID), nil
	case "openai/gpt-image-1":
		return getGPTImageSchema(), nil
	case "openai/gpt-image-1.5":
		return getGPTImage15Schema(), nil
	default:
		return nil, fmt.Errorf("unknown model ID: %s", modelID)
	}
}

func getGPTImageSchema() *ModelConfigSchema {
	return &ModelConfigSchema{
		ModelID:     "openai/gpt-image-1",
		DisplayName: "OpenAI GPT Image 1",
		Fields: []ModelConfigField{
			{
				Name:        "openai_api_key",
				Type:        "string",
				Default:     "",
				Description: "Your OpenAI API key",
				Required:    true,
			},
			{
				Name:        "prompt",
				Type:        "string",
				Default:     "",
				Description: "A text description of the desired image",
				Required:    true,
			},
			{
				Name:        "quality",
				Type:        "string",
				Default:     "auto",
				Description: "The quality of the generated image",
				Options:     []string{"low", "medium", "high", "auto"},
				Required:    true,
			},
			{
				Name:    "user_id",
				Type:    "string",
				Default: "",
				Description: "An optional unique identifier representing your end-user. " +
					"This helps OpenAI monitor and detect abuse.",
			},
			{
				Name:        "background",
				Type:        "string",
				Default:     "auto",
				Description: "Set whether the background is transparent or opaque or choose automatically",
				Options:     []string{"auto", "transparent", "opaque"},
				Required:    true,
			},
			{
				Name:        "moderation",
				Type:        "string",
				Default:     "auto",
				Description: "Content moderation level",
				Options:     []string{"auto", "low"},
				Required:    true,
			},
			{
				Name:        "aspect_ratio",
				Type:        "string",
				Default:     "1:1",
				Description: "The aspect ratio of the generated image",
				Options:     []string{"1:1", "3:2", "2:3"},
				Required:    true,
			},
			{
				Name:        "number_of_images",
				Type:        "int",
				Default:     1,
				Description: "Number of images to generate (1-10)",
				Min:         ptr(1.0),
				Max:         ptr(10.0),
				Required:    true,
			},
			{
				Name:        "output_compression",
				Type:        "int",
				Default:     90,
				Description: "Compression level (0-100%)",
				Min:         ptr(0.0),
				Max:         ptr(100.0),
			},
			{
				Name:        "output_format",
				Type:        "string",
				Default:     "webp",
				Description: "Output image format",
				Options:     []string{"png", "jpeg", "webp"},
				Required:    true,
			},
			{
				Name:        "input_fidelity",
				Type:        "string",
				Default:     "low",
				Description: "Input image fidelity level",
				Options:     []string{"low", "high"},
				Required:    true,
			},
		},
	}
}

func getGPTImage15Schema() *ModelConfigSchema {
	return &ModelConfigSchema{
		ModelID:     "openai/gpt-image-1.5",
		DisplayName: "OpenAI GPT Image 1.5",
		Fields: []ModelConfigField{
			{
				Name:        "openai_api_key",
				Type:        "string",
				Default:     "",
				Description: "Your OpenAI API key",
				Required:    true,
			},
			{
				Name:        "prompt",
				Type:        "string",
				Default:     "",
				Description: "A text description of the desired image",
				Required:    true,
			},
			{
				Name:        "quality",
				Type:        "string",
				Default:     "auto",
				Description: "The quality of the generated image",
				Options:     []string{"low", "medium", "high", "auto"},
				Required:    true,
			},
			{
				Name:    "user_id",
				Type:    "string",
				Default: "",
				Description: "An optional unique identifier representing your end-user. " +
					"This helps OpenAI monitor and detect abuse.",
			},
			{
				Name:        "background",
				Type:        "string",
				Default:     "auto",
				Description: "Set whether the background is transparent or opaque or choose automatically",
				Options:     []string{"auto", "transparent", "opaque"},
				Required:    true,
			},
			{
				Name:        "moderation",
				Type:        "string",
				Default:     "auto",
				Description: "Content moderation level",
				Options:     []string{"auto", "low"},
				Required:    true,
			},
			{
				Name:        "aspect_ratio",
				Type:        "string",
				Default:     "1:1",
				Description: "The aspect ratio of the generated image",
				Options:     []string{"1:1", "3:2", "2:3"},
				Required:    true,
			},
			{
				Name:        "number_of_images",
				Type:        "int",
				Default:     1,
				Description: "Number of images to generate (1-10)",
				Min:         ptr(1.0),
				Max:         ptr(10.0),
				Required:    true,
			},
			{
				Name:        "output_compression",
				Type:        "int",
				Default:     90,
				Description: "Compression level (0-100%)",
				Min:         ptr(0.0),
				Max:         ptr(100.0),
			},
			{
				Name:        "output_format",
				Type:        "string",
				Default:     "webp",
				Description: "Output image format",
				Options:     []string{"png", "jpeg", "webp"},
				Required:    true,
			},
			{
				Name:        "input_fidelity",
				Type:        "string",
				Default:     "low",
				Description: "Input image fidelity level",
				Options:     []string{"low", "high"},
				Required:    true,
			},
		},
	}
}

func getQwenSchema() *ModelConfigSchema {
	return &ModelConfigSchema{
		ModelID:     "qwen/qwen-image-edit",
		DisplayName: "Qwen Image Edit",
		Fields: []ModelConfigField{
			{
				Name:        "go_fast",
				Type:        "bool",
				Default:     true,
				Description: "Enable fast mode for quicker processing",
				Required:    true,
			},
			{
				Name:        "aspect_ratio",
				Type:        "string",
				Default:     "match_input_image",
				Description: "Output aspect ratio",
				Options:     []string{"1:1", "16:9", "4:3", "3:2", "match_input_image"},
				Required:    true,
			},
			{
				Name:        "output_format",
				Type:        "string",
				Default:     "webp",
				Description: "Output image format",
				Options:     []string{"webp", "png", "jpg"},
				Required:    true,
			},
			{
				Name:        "output_quality",
				Type:        "int",
				Default:     80,
				Description: "Output image quality (1-100)",
				Min:         ptr(1.0),
				Max:         ptr(100.0),
				Required:    true,
			},
		},
	}
}

func getFluxKontextSchema(modelID string) *ModelConfigSchema {
	displayName := "Flux Kontext Max"
	if modelID == "black-forest-labs/flux-kontext-pro" {
		displayName = "Flux Kontext Pro"
	}

	return &ModelConfigSchema{
		ModelID:     modelID,
		DisplayName: displayName,
		Fields: []ModelConfigField{
			{
				Name:        "aspect_ratio",
				Type:        "string",
				Default:     "match_input_image",
				Description: "Output aspect ratio",
				Options:     []string{"1:1", "16:9", "4:3", "3:2", "match_input_image"},
				Required:    true,
			},
			{
				Name:        "output_format",
				Type:        "string",
				Default:     "png",
				Description: "Output image format",
				Options:     []string{"webp", "png", "jpg"},
				Required:    true,
			},
			{
				Name:        "safety_tolerance",
				Type:        "int",
				Default:     4,
				Description: "Safety filter tolerance (1=strict, 6=permissive)",
				Min:         ptr(1.0),
				Max:         ptr(6.0),
				Required:    true,
			},
			{
				Name:        "prompt_upsampling",
				Type:        "bool",
				Default:     false,
				Description: "Enhance prompts automatically",
				Required:    true,
			},
			{
				Name:        "num_outputs",
				Type:        "int",
				Default:     1,
				Description: "Number of images to generate",
				Min:         ptr(1.0),
				Max:         ptr(4.0),
				Required:    true,
			},
			{
				Name:        "output_quality",
				Type:        "int",
				Default:     90,
				Description: "Output image quality (1-100)",
				Min:         ptr(1.0),
				Max:         ptr(100.0),
				Required:    true,
			},
		},
	}
}

func getSeedreamSchema(modelID string) *ModelConfigSchema {
	displayName := "Seedream 3"
	if modelID == "bytedance/seedream-4" {
		displayName = "Seedream 4"
	}

	return &ModelConfigSchema{
		ModelID:     modelID,
		DisplayName: displayName,
		Fields: []ModelConfigField{
			{
				Name:        "aspect_ratio",
				Type:        "string",
				Default:     "1:1",
				Description: "Output aspect ratio",
				Options:     []string{"1:1", "16:9", "4:3", "3:2"},
				Required:    true,
			},
			{
				Name:        "num_inference_steps",
				Type:        "int",
				Default:     50,
				Description: "Number of denoising steps (more = higher quality, slower)",
				Min:         ptr(20.0),
				Max:         ptr(100.0),
				Required:    true,
			},
			{
				Name:        "guidance_scale",
				Type:        "float",
				Default:     7.5,
				Description: "How closely to follow the prompt (1.0-20.0)",
				Min:         ptr(1.0),
				Max:         ptr(20.0),
				Required:    true,
			},
			{
				Name:        "output_quality",
				Type:        "int",
				Default:     95,
				Description: "Output image quality (1-100)",
				Min:         ptr(1.0),
				Max:         ptr(100.0),
				Required:    true,
			},
		},
	}
}

func ptr(f float64) *float64 {
	return &f
}
