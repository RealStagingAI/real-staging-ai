package model

import (
	"encoding/json"
	"fmt"
)

// ModelConfig defines the interface for model-specific configuration.
type ModelConfig interface {
	// ToMap converts the config to a map for Replicate API
	ToMap() map[string]interface{}

	// Validate checks if the configuration is valid
	Validate() error

	// GetDefaults returns a new instance with default values
	GetDefaults() ModelConfig
}

// ConfigField defines metadata for a configuration field (used for UI generation).
type ConfigField struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // "string", "int", "float", "bool"
	Default     interface{} `json:"default"`
	Description string      `json:"description"`
	Options     []string    `json:"options,omitempty"` // For dropdown fields
	Min         *float64    `json:"min,omitempty"`     // For numeric fields
	Max         *float64    `json:"max,omitempty"`     // For numeric fields
	Required    bool        `json:"required"`
}

// ConfigSchema describes the configuration structure for UI generation.
type ConfigSchema struct {
	ModelID     ModelID       `json:"model_id"`
	DisplayName string        `json:"display_name"`
	Fields      []ConfigField `json:"fields"`
}

// QwenConfig contains all Qwen Image Edit model parameters.
type QwenConfig struct {
	GoFast        bool   `json:"go_fast"`
	AspectRatio   string `json:"aspect_ratio"`
	OutputFormat  string `json:"output_format"`
	OutputQuality int    `json:"output_quality"`
	Seed          *int64 `json:"seed,omitempty"`
}

// ToMap converts QwenConfig to a map for Replicate API.
func (c *QwenConfig) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"go_fast":        c.GoFast,
		"aspect_ratio":   c.AspectRatio,
		"output_format":  c.OutputFormat,
		"output_quality": c.OutputQuality,
	}
	if c.Seed != nil {
		m["seed"] = *c.Seed
	}
	return m
}

// Validate checks if QwenConfig is valid.
func (c *QwenConfig) Validate() error {
	validAspectRatios := []string{"1:1", "16:9", "4:3", "3:2", "match_input_image"}
	if !contains(validAspectRatios, c.AspectRatio) {
		return fmt.Errorf("invalid aspect_ratio: %s", c.AspectRatio)
	}

	validFormats := []string{"webp", "png", "jpg"}
	if !contains(validFormats, c.OutputFormat) {
		return fmt.Errorf("invalid output_format: %s", c.OutputFormat)
	}

	if c.OutputQuality < 1 || c.OutputQuality > 100 {
		return fmt.Errorf("output_quality must be between 1 and 100, got %d", c.OutputQuality)
	}

	return nil
}

// GetDefaults returns QwenConfig with default values.
func (c *QwenConfig) GetDefaults() ModelConfig {
	return &QwenConfig{
		GoFast:        true,
		AspectRatio:   "match_input_image",
		OutputFormat:  "webp",
		OutputQuality: 80,
	}
}

// FluxKontextConfig contains all Flux Kontext model parameters.
type FluxKontextConfig struct {
	AspectRatio      string `json:"aspect_ratio"`
	OutputFormat     string `json:"output_format"`
	SafetyTolerance  int    `json:"safety_tolerance"`
	PromptUpsampling bool   `json:"prompt_upsampling"`
	NumOutputs       int    `json:"num_outputs"`
	OutputQuality    int    `json:"output_quality"`
	Seed             *int64 `json:"seed,omitempty"`
}

// ToMap converts FluxKontextConfig to a map for Replicate API.
func (c *FluxKontextConfig) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"aspect_ratio":      c.AspectRatio,
		"output_format":     c.OutputFormat,
		"safety_tolerance":  c.SafetyTolerance,
		"prompt_upsampling": c.PromptUpsampling,
		"num_outputs":       c.NumOutputs,
		"output_quality":    c.OutputQuality,
	}
	if c.Seed != nil {
		m["seed"] = *c.Seed
	}
	return m
}

// Validate checks if FluxKontextConfig is valid.
func (c *FluxKontextConfig) Validate() error {
	validAspectRatios := []string{"1:1", "16:9", "4:3", "3:2", "match_input_image"}
	if !contains(validAspectRatios, c.AspectRatio) {
		return fmt.Errorf("invalid aspect_ratio: %s", c.AspectRatio)
	}

	validFormats := []string{"webp", "png", "jpg"}
	if !contains(validFormats, c.OutputFormat) {
		return fmt.Errorf("invalid output_format: %s", c.OutputFormat)
	}

	if c.SafetyTolerance < 1 || c.SafetyTolerance > 6 {
		return fmt.Errorf("safety_tolerance must be between 1 and 6, got %d", c.SafetyTolerance)
	}

	if c.NumOutputs < 1 || c.NumOutputs > 4 {
		return fmt.Errorf("num_outputs must be between 1 and 4, got %d", c.NumOutputs)
	}

	if c.OutputQuality < 1 || c.OutputQuality > 100 {
		return fmt.Errorf("output_quality must be between 1 and 100, got %d", c.OutputQuality)
	}

	return nil
}

// GetDefaults returns FluxKontextConfig with default values.
func (c *FluxKontextConfig) GetDefaults() ModelConfig {
	return &FluxKontextConfig{
		AspectRatio:      "match_input_image",
		OutputFormat:     "png",
		SafetyTolerance:  4,
		PromptUpsampling: false,
		NumOutputs:       1,
		OutputQuality:    90,
	}
}

// SeedreamConfig contains all Seedream model parameters.
type SeedreamConfig struct {
	AspectRatio       string  `json:"aspect_ratio"`
	NumInferenceSteps int     `json:"num_inference_steps"`
	GuidanceScale     float64 `json:"guidance_scale"`
	OutputQuality     int     `json:"output_quality"`
	Seed              *int64  `json:"seed,omitempty"`
}

// ToMap converts SeedreamConfig to a map for Replicate API.
func (c *SeedreamConfig) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"aspect_ratio":        c.AspectRatio,
		"num_inference_steps": c.NumInferenceSteps,
		"guidance_scale":      c.GuidanceScale,
		"output_quality":      c.OutputQuality,
	}
	if c.Seed != nil {
		m["seed"] = *c.Seed
	}
	return m
}

// Validate checks if SeedreamConfig is valid.
func (c *SeedreamConfig) Validate() error {
	validAspectRatios := []string{"1:1", "16:9", "4:3", "3:2"}
	if !contains(validAspectRatios, c.AspectRatio) {
		return fmt.Errorf("invalid aspect_ratio: %s", c.AspectRatio)
	}

	if c.NumInferenceSteps < 20 || c.NumInferenceSteps > 100 {
		return fmt.Errorf("num_inference_steps must be between 20 and 100, got %d", c.NumInferenceSteps)
	}

	if c.GuidanceScale < 1.0 || c.GuidanceScale > 20.0 {
		return fmt.Errorf("guidance_scale must be between 1.0 and 20.0, got %f", c.GuidanceScale)
	}

	if c.OutputQuality < 1 || c.OutputQuality > 100 {
		return fmt.Errorf("output_quality must be between 1 and 100, got %d", c.OutputQuality)
	}

	return nil
}

// GetDefaults returns SeedreamConfig with default values.
func (c *SeedreamConfig) GetDefaults() ModelConfig {
	return &SeedreamConfig{
		AspectRatio:       "1:1",
		NumInferenceSteps: 50,
		GuidanceScale:     7.5,
		OutputQuality:     95,
	}
}

// ParseModelConfig parses JSON into the appropriate ModelConfig type.
func ParseModelConfig(modelID ModelID, data []byte) (ModelConfig, error) {
	var config ModelConfig

	switch modelID {
	case ModelQwenImageEdit:
		config = &QwenConfig{}
	case ModelFluxKontextMax, ModelFluxKontextPro:
		config = &FluxKontextConfig{}
	case ModelSeedream3, ModelSeedream4:
		config = &SeedreamConfig{}
	default:
		return nil, fmt.Errorf("unknown model ID: %s", modelID)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return config, nil
}

// GetConfigSchema returns the schema for a given model configuration.
func GetConfigSchema(modelID ModelID) (*ConfigSchema, error) {
	switch modelID {
	case ModelQwenImageEdit:
		return getQwenSchema(), nil
	case ModelFluxKontextMax, ModelFluxKontextPro:
		return getFluxKontextSchema(modelID), nil
	case ModelSeedream3, ModelSeedream4:
		return getSeedreamSchema(modelID), nil
	default:
		return nil, fmt.Errorf("unknown model ID: %s", modelID)
	}
}

func getQwenSchema() *ConfigSchema {
	return &ConfigSchema{
		ModelID:     ModelQwenImageEdit,
		DisplayName: "Qwen Image Edit",
		Fields: []ConfigField{
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

func getFluxKontextSchema(modelID ModelID) *ConfigSchema {
	displayName := "Flux Kontext Max"
	if modelID == ModelFluxKontextPro {
		displayName = "Flux Kontext Pro"
	}

	return &ConfigSchema{
		ModelID:     modelID,
		DisplayName: displayName,
		Fields: []ConfigField{
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

func getSeedreamSchema(modelID ModelID) *ConfigSchema {
	displayName := "Seedream 3"
	if modelID == ModelSeedream4 {
		displayName = "Seedream 4"
	}

	return &ConfigSchema{
		ModelID:     modelID,
		DisplayName: displayName,
		Fields: []ConfigField{
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

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func ptr(f float64) *float64 {
	return &f
}
