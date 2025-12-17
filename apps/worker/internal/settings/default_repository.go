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
	configKeyGPTImage1      = "gpt_image_1"
	configKeyGPTImage1_5    = "gpt_image_1_5"
)

// DefaultRepository provides access to settings stored in the database.
type DefaultRepository struct {
	db *sql.DB
}

// NewDefaultRepository creates a new settings repository.
func NewDefaultRepository(db *sql.DB) *DefaultRepository {
	return &DefaultRepository{db: db}
}

// GetActiveModel retrieves the active model ID from settings.
// Returns the default model if not found.
func (r *DefaultRepository) GetActiveModel(ctx context.Context) (model.ID, error) {
	var value string
	query := `SELECT value FROM settings WHERE key = $1`

	err := r.db.QueryRowContext(ctx, query, "active_model").Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return default model if setting doesn't exist
			return model.ModelFluxKontextMax, nil
		}
		return "", fmt.Errorf("failed to query active model: %w", err)
	}

	// Validate the model ID exists in registry
	modelID := model.ID(value)
	return modelID, nil
}

// GetModelConfig retrieves the configuration for a specific model.
func (r *DefaultRepository) GetModelConfig(
	ctx context.Context, modelID model.ID,
) (model.Config, error) {
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
func (r *DefaultRepository) UpdateModelConfig(
	ctx context.Context, modelID model.ID, config model.Config, userID string,
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
func getConfigKey(modelID model.ID) string {
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
	case model.ModelGPTImage1:
		return configKeyGPTImage1
	case model.ModelGPTImage1_5:
		return configKeyGPTImage1_5
	default:
		return string(modelID)
	}
}
