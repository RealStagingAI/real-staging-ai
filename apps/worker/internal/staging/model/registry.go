package model

import (
	"context"
	"fmt"

	"github.com/replicate/replicate-go"
)

// ID represents a unique identifier for a supported AI model.
type ID string

// Supported models
const (
	ModelQwenImageEdit  ID = "qwen/qwen-image-edit"
	ModelFluxKontextMax ID = "black-forest-labs/flux-kontext-max"
	ModelFluxKontextPro ID = "black-forest-labs/flux-kontext-pro"
	ModelSeedream3      ID = "bytedance/seedream-3"
	ModelSeedream4      ID = "bytedance/seedream-4"
	ModelGPTImage1      ID = "openai/gpt-image-1"
)

// ModelInputRequest contains the parameters needed to build model input.
type ModelInputRequest struct {
	ImageDataURL string
	Prompt       string
	Seed         *int64
	Config       Config // Optional: model-specific configuration (uses defaults if nil)
}

// ModelInputBuilder defines the interface for building model-specific input parameters.
//
//go:generate go run github.com/matryer/moq@v0.5.3 -out model_mock.go . ModelInputBuilder
type ModelInputBuilder interface {
	// BuildInput creates the input parameters for the model.
	BuildInput(ctx context.Context, req *ModelInputRequest) (replicate.PredictionInput, error)

	// Validate checks if the request is valid for this model.
	Validate(req *ModelInputRequest) error
}

// ModelMetadata contains information about a registered model.
type ModelMetadata struct {
	ID            ID
	Name          string
	Description   string
	Version       string
	InputBuilder  ModelInputBuilder
	DefaultConfig Config // Default configuration for this model
}

// ModelRegistry manages the available AI models and their configurations.
type ModelRegistry struct {
	models map[ID]*ModelMetadata
}

// NewModelRegistry creates a new registry with all supported models.
func NewModelRegistry() *ModelRegistry {
	registry := &ModelRegistry{
		models: make(map[ID]*ModelMetadata),
	}

	// Register Qwen Image Edit model
	registry.Register(&ModelMetadata{
		ID:            ModelQwenImageEdit,
		Name:          "Qwen Image Edit",
		Description:   "Fast image editing model optimized for virtual staging",
		Version:       "latest",
		InputBuilder:  NewQwenInputBuilder(),
		DefaultConfig: (&QwenConfig{}).GetDefaults(),
	})

	// Register Flux Kontext Max model
	registry.Register(&ModelMetadata{
		ID:            ModelFluxKontextMax,
		Name:          "Flux Kontext Max",
		Description:   "High-quality image generation and editing with advanced context understanding",
		Version:       "latest",
		InputBuilder:  NewFluxKontextInputBuilder(),
		DefaultConfig: (&FluxKontextConfig{}).GetDefaults(),
	})

	// Register Flux Kontext Pro model
	registry.Register(&ModelMetadata{
		ID:            ModelFluxKontextPro,
		Name:          "Flux Kontext Pro",
		Description:   "State-of-the-art text-based image editing with high-quality outputs and excellent prompt following",
		Version:       "latest",
		InputBuilder:  NewFluxKontextInputBuilder(),
		DefaultConfig: (&FluxKontextConfig{}).GetDefaults(),
	})

	// Register Seedream models
	registry.Register(&ModelMetadata{
		ID:            ModelSeedream3,
		Name:          "Seedream 3",
		Description:   "Unified text-to-image generation and precise editing",
		Version:       "latest",
		InputBuilder:  NewSeedreamInputBuilder(),
		DefaultConfig: (&SeedreamConfig{}).GetDefaults(),
	})

	registry.Register(&ModelMetadata{
		ID:            ModelSeedream4,
		Name:          "Seedream 4",
		Description:   "Unified text-to-image generation and precise editing at up to 4K resolution",
		Version:       "latest",
		InputBuilder:  NewSeedreamInputBuilder(),
		DefaultConfig: (&SeedreamConfig{}).GetDefaults(),
	})

	// Register GPT Image 1 model
	registry.Register(&ModelMetadata{
		ID:            ModelGPTImage1,
		Name:          "GPT Image 1",
		Description:   "OpenAI's GPT Image 1 model providing multimodal image generation",
		Version:       "5ac56c15446a60fa63b3823de926ada90f5971c2cf9b1dd07659126cfda434e6",
		InputBuilder:  NewGPTImageInputBuilder(),
		DefaultConfig: (&GPTImageConfig{}).GetDefaults(),
	})

	return registry
}

// Register adds a model to the registry.
func (r *ModelRegistry) Register(metadata *ModelMetadata) {
	r.models[metadata.ID] = metadata
}

// Get retrieves a model's metadata by ID.
func (r *ModelRegistry) Get(id ID) (*ModelMetadata, error) {
	model, exists := r.models[id]
	if !exists {
		return nil, fmt.Errorf("model not found: %s", id)
	}
	return model, nil
}

// List returns all registered models.
func (r *ModelRegistry) List() []*ModelMetadata {
	models := make([]*ModelMetadata, 0, len(r.models))
	for _, model := range r.models {
		models = append(models, model)
	}
	return models
}

// Exists checks if a model is registered.
func (r *ModelRegistry) Exists(id ID) bool {
	_, exists := r.models[id]
	return exists
}
