package settings

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/real-staging-ai/worker/internal/staging/model"
)

// Repository provides access to settings stored in the database.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new settings repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetActiveModel retrieves the active model ID from settings.
// Returns the default model if not found.
func (r *Repository) GetActiveModel(ctx context.Context) (model.ModelID, error) {
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
	modelID := model.ModelID(value)
	return modelID, nil
}
