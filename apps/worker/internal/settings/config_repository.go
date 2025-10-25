package settings

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/real-staging-ai/worker/internal/staging/model"
)

const (
	configKeyQwen           = "qwen"
	configKeyFluxKontextMax = "flux_kontext_max"
	configKeyFluxKontextPro = "flux_kontext_pro"
	configKeySeedream3      = "seedream_3"
	configKeySeedream4      = "seedream_4"
)

// ConfigRepository handles model configuration persistence.
type ConfigRepository interface {
	// GetModelConfig retrieves the configuration for a specific model
	GetModelConfig(ctx context.Context, modelID model.ModelID) (model.ModelConfig, error)

	// UpdateModelConfig updates the configuration for a specific model
	UpdateModelConfig(ctx context.Context, modelID model.ModelID, config model.ModelConfig, userID string) error
}

// DefaultConfigRepository implements ConfigRepository using PostgreSQL.
type DefaultConfigRepository struct {
	db *sql.DB
}

// NewConfigRepository creates a new DefaultConfigRepository.
func NewConfigRepository(db *sql.DB) *DefaultConfigRepository {
	return &DefaultConfigRepository{db: db}
}

// GetModelConfig retrieves the configuration for a specific model.
func (r *DefaultConfigRepository) GetModelConfig(
	ctx context.Context, modelID model.ModelID,
) (model.ModelConfig, error) {
	key := fmt.Sprintf("model_config_%s", getConfigKey(modelID))

	query := `SELECT model_settings FROM settings WHERE key = $1`
	var configJSON []byte
	err := r.db.QueryRowContext(ctx, query, key).Scan(&configJSON)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("config not found for model: %s", modelID)
		}
		return nil, fmt.Errorf("failed to query config: %w", err)
	}

	config, err := model.ParseModelConfig(modelID, configJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return config, nil
}

// UpdateModelConfig updates the configuration for a specific model.
func (r *DefaultConfigRepository) UpdateModelConfig(
	ctx context.Context, modelID model.ModelID, config model.ModelConfig, userID string,
) error {
	// Validate config before storing
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	key := fmt.Sprintf("model_config_%s", getConfigKey(modelID))

	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	query := `
		UPDATE settings 
		SET model_settings = $1, updated_at = NOW(), updated_by = $2
		WHERE key = $3
	`
	result, err := r.db.ExecContext(ctx, query, configJSON, userID, key)
	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("config not found for model: %s", modelID)
	}

	return nil
}

// getConfigKey converts a ModelID to its configuration key suffix.
func getConfigKey(modelID model.ModelID) string {
	switch modelID {
	case model.ModelQwenImageEdit:
		return configKeyQwen
	case model.ModelFluxKontextMax:
		return configKeyFluxKontextMax
	case model.ModelFluxKontextPro:
		return configKeyFluxKontextPro
	case model.ModelSeedream3:
		return configKeySeedream3
	case model.ModelSeedream4:
		return configKeySeedream4
	default:
		return string(modelID)
	}
}
