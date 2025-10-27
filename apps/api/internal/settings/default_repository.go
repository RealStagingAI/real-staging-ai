package settings

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/real-staging-ai/api/internal/storage"
)

// DefaultRepository implements Repository using PostgreSQL.
type DefaultRepository struct {
	db storage.PgxPool
}

// Ensure DefaultRepository implements Repository.
var _ Repository = (*DefaultRepository)(nil)

// NewDefaultRepository creates a new DefaultRepository.
func NewDefaultRepository(db storage.PgxPool) *DefaultRepository {
	return &DefaultRepository{db: db}
}

// GetByKey retrieves a setting by its key.
func (r *DefaultRepository) GetByKey(ctx context.Context, key string) (*Setting, error) {
	query := `
		SELECT key, value, description, updated_at, updated_by
		FROM settings
		WHERE key = $1
	`

	var setting Setting
	var updatedBy *string

	err := r.db.QueryRow(ctx, query, key).Scan(
		&setting.Key,
		&setting.Value,
		&setting.Description,
		&setting.UpdatedAt,
		&updatedBy,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("setting not found: %s", key)
		}
		return nil, fmt.Errorf("failed to get setting: %w", err)
	}

	setting.UpdatedBy = updatedBy

	return &setting, nil
}

// Update updates a setting value.
func (r *DefaultRepository) Update(ctx context.Context, key, value, userID string) error {
	query := `
		UPDATE settings
		SET value = $1, updated_at = NOW(), updated_by = $2
		WHERE key = $3
	`

	result, err := r.db.Exec(ctx, query, value, userID, key)
	if err != nil {
		return fmt.Errorf("failed to update setting: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("setting not found: %s", key)
	}

	return nil
}

// List retrieves all settings.
func (r *DefaultRepository) List(ctx context.Context) ([]Setting, error) {
	query := `
		SELECT key, value, description, updated_at, updated_by
		FROM settings
		ORDER BY key
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list settings: %w", err)
	}
	defer rows.Close()

	var settings []Setting
	for rows.Next() {
		var setting Setting
		var updatedBy *string

		err := rows.Scan(
			&setting.Key,
			&setting.Value,
			&setting.Description,
			&setting.UpdatedAt,
			&updatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan setting: %w", err)
		}

		setting.UpdatedBy = updatedBy
		settings = append(settings, setting)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating settings: %w", err)
	}

	return settings, nil
}

// GetModelConfig retrieves the configuration JSON for a specific model.
func (r *DefaultRepository) GetModelConfig(ctx context.Context, modelID string) ([]byte, error) {
	key := getConfigKey(modelID)

	query := `SELECT model_settings FROM settings WHERE key = $1`
	var configJSON []byte
	err := r.db.QueryRow(ctx, query, key).Scan(&configJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("config not found for model: %s", modelID)
		}
		return nil, fmt.Errorf("failed to query config: %w", err)
	}

	return configJSON, nil
}

// UpdateModelConfig updates the configuration JSON for a specific model.
func (r *DefaultRepository) UpdateModelConfig(
	ctx context.Context, modelID string, configJSON []byte, userID string,
) error {
	key := getConfigKey(modelID)

	query := `
		UPDATE settings 
		SET model_settings = $1, updated_at = NOW(), updated_by = $2
		WHERE key = $3
	`
	result, err := r.db.Exec(ctx, query, configJSON, userID, key)
	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("config not found for model: %s", modelID)
	}

	return nil
}

// getConfigKey converts a model ID to its configuration key.
func getConfigKey(modelID string) string {
	// Convert model ID to config key format
	// e.g., "qwen/qwen-image-edit" -> "model_config_qwen"
	switch modelID {
	case "qwen/qwen-image-edit":
		return "model_config_qwen"
	case "black-forest-labs/flux-kontext-max":
		return "model_config_flux_kontext_max"
	case "black-forest-labs/flux-kontext-pro":
		return "model_config_flux_kontext_pro"
	case "bytedance/seedream-3":
		return "model_config_seedream_3"
	case "bytedance/seedream-4":
		return "model_config_seedream_4"
	case "openai/gpt-image-1":
		return "model_config_gpt_image_1"
	default:
		return fmt.Sprintf("model_config_%s", modelID)
	}
}
