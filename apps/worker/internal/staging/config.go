package staging

import (
	"context"

	"github.com/real-staging-ai/worker/internal/staging/model"
)

// ConfigRepository handles model configuration persistence.
type ConfigRepository interface {
	// GetModelConfig retrieves the configuration for a specific model
	GetModelConfig(ctx context.Context, modelID model.ModelID) (model.ModelConfig, error)
}
