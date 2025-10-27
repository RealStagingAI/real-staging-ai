package settings

import (
	"context"

	"github.com/real-staging-ai/worker/internal/staging/model"
)

//go:generate go run github.com/matryer/moq@v0.5.3 -out repository_mock.go . Repository

type Repository interface {
	// GetActiveModel retrieves the active model ID from settings.
	// Returns the default model if not found.
	GetActiveModel(ctx context.Context) (model.ID, error)

	// GetModelConfig retrieves the configuration for a specific model
	GetModelConfig(ctx context.Context, modelID model.ID) (model.Config, error)

	// UpdateModelConfig updates the configuration for a specific model
	UpdateModelConfig(ctx context.Context, modelID model.ID, config model.Config, userID string) error
}
