package settings

import "time"

// Setting represents a system configuration setting.
type Setting struct {
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Description *string   `json:"description,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
	UpdatedBy   *string   `json:"updated_by,omitempty"`
}

// ModelInfo represents information about an available AI model.
type ModelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	IsActive    bool   `json:"is_active"`
}

// UpdateSettingRequest represents a request to update a setting.
type UpdateSettingRequest struct {
	Value string `json:"value"`
}

// ModelConfigField represents metadata for a configuration field.
type ModelConfigField struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Default     interface{} `json:"default"`
	Description string      `json:"description"`
	Options     []string    `json:"options,omitempty"`
	Min         *float64    `json:"min,omitempty"`
	Max         *float64    `json:"max,omitempty"`
	Required    bool        `json:"required"`
}

// ModelConfigSchema describes the configuration structure for a model.
type ModelConfigSchema struct {
	ModelID     string             `json:"model_id"`
	DisplayName string             `json:"display_name"`
	Fields      []ModelConfigField `json:"fields"`
}

// ModelConfig represents the stored configuration for a model.
type ModelConfig struct {
	ModelID string                 `json:"model_id"`
	Config  map[string]interface{} `json:"config"`
}
