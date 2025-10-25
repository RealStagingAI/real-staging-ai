package admin

import "github.com/labstack/echo/v4"

//go:generate go run github.com/matryer/moq@v0.5.3 -out handler_mock.go . Handler

type Handler interface {
	// ListModels handles GET /admin/models - Lists all available AI models.
	ListModels(c echo.Context) error

	// GetActiveModel handles GET /admin/models/active - Gets the currently active model.
	GetActiveModel(c echo.Context) error

	// UpdateActiveModel handles PUT /admin/models/active - Updates the active model.
	UpdateActiveModel(c echo.Context) error

	// ListSettings handles GET /admin/settings - Lists all settings.
	ListSettings(c echo.Context) error

	// GetSetting handles GET /admin/settings/:key - Gets a specific setting.
	GetSetting(c echo.Context) error

	// UpdateSetting handles PUT /admin/settings/:key - Updates a setting.
	UpdateSetting(c echo.Context) error

	// GetModelConfig handles GET /admin/models/:id/config - Gets the configuration for a model.
	GetModelConfig(c echo.Context) error

	// UpdateModelConfig handles PUT /admin/models/:id/config - Updates the configuration for a model.
	UpdateModelConfig(c echo.Context) error

	// GetModelConfigSchema handles GET /admin/models/:id/config/schema - Gets the schema for a model's configuration.
	GetModelConfigSchema(c echo.Context) error

	// resolveUserUUID looks up or creates a user based on Auth0 sub, returning the user's UUID.
	resolveUserUUID(c echo.Context) (string, error)
}
