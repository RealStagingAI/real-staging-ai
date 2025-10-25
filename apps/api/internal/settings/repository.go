package settings

import (
	"context"
)

//go:generate go run github.com/matryer/moq@v0.5.3 -out repository_mock.go . Repository

// Repository defines the interface for settings data access.
type Repository interface {
	// GetByKey retrieves a setting by its key.
	GetByKey(ctx context.Context, key string) (*Setting, error)

	// Update updates a setting value.
	Update(ctx context.Context, key, value, userID string) error

	// List retrieves all settings.
	List(ctx context.Context) ([]Setting, error)

	// GetModelConfig retrieves the configuration JSON for a specific model.
	GetModelConfig(ctx context.Context, modelID string) ([]byte, error)

	// UpdateModelConfig updates the configuration JSON for a specific model.
	UpdateModelConfig(ctx context.Context, modelID string, configJSON []byte, userID string) error
}
