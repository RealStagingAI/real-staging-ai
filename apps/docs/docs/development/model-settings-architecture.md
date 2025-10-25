# Model Settings Architecture

## Overview

This document describes the architecture for exposing and managing model-specific settings.

## Goals

1. **Expose all API parameters** for each model (Qwen, Flux Kontext, Seedream, etc.)
2. **Structured configuration** using Go structs instead of hardcoded values
3. **Database-backed settings** allowing runtime configuration without code changes
4. **Dynamic UI generation** in the admin panel based on model configuration
5. **Backward compatibility** with existing prompt library system

## Architecture

### 1. Model Configuration Structs

Each model will have a dedicated configuration struct containing all API parameters:

```go
// apps/worker/internal/staging/model/config.go

// ModelConfig defines the interface for model-specific configuration
type ModelConfig interface {
    // ToMap converts the config to a map for Replicate API
    ToMap() map[string]interface{}
    
    // Validate checks if the configuration is valid
    Validate() error
    
    // Schema returns the JSON schema for UI generation
    Schema() ConfigSchema
}

// QwenConfig contains all Qwen-specific parameters
type QwenConfig struct {
    GoFast        bool   `json:"go_fast" default:"true" description:"Enable fast mode"`
    AspectRatio   string `json:"aspect_ratio" default:"match_input_image" options:"1:1,16:9,4:3,match_input_image"`
    OutputFormat  string `json:"output_format" default:"webp" options:"webp,png,jpg"`
    OutputQuality int    `json:"output_quality" default:"80" min:"1" max:"100"`
    Seed          *int64 `json:"seed,omitempty" description:"Random seed for reproducibility"`
}

// FluxKontextConfig contains all Flux Kontext parameters
type FluxKontextConfig struct {
    AspectRatio       string  `json:"aspect_ratio" default:"match_input_image" options:"1:1,16:9,4:3,match_input_image"`
    OutputFormat      string  `json:"output_format" default:"png" options:"webp,png,jpg"`
    SafetyTolerance   int     `json:"safety_tolerance" default:"2" min:"1" max:"6"`
    PromptUpsampling  bool    `json:"prompt_upsampling" default:"false"`
    Seed              *int64  `json:"seed,omitempty"`
    NumOutputs        int     `json:"num_outputs" default:"1" min:"1" max:"4"`
    OutputQuality     int     `json:"output_quality" default:"90" min:"1" max:"100"`
}

// SeedreamConfig contains all Seedream parameters  
type SeedreamConfig struct {
    AspectRatio      string  `json:"aspect_ratio" default:"1:1" options:"1:1,16:9,4:3,3:2"`
    Seed             *int64  `json:"seed,omitempty"`
    NumInferenceSteps int    `json:"num_inference_steps" default:"50" min:"20" max:"100"`
    GuidanceScale    float64 `json:"guidance_scale" default:"7.5" min:"1.0" max:"20.0"`
    OutputQuality    int     `json:"output_quality" default:"95" min:"1" max:"100"`
}
```

### 2. Database Schema

Extend the settings table to store model-specific JSON configurations:

```sql
-- New migration: 0009_add_model_settings.up.sql

-- Add model_settings column to existing settings table
ALTER TABLE settings 
ADD COLUMN model_settings JSONB DEFAULT '{}'::jsonb;

-- Create index for faster JSON queries
CREATE INDEX idx_settings_model_settings ON settings USING gin(model_settings);

-- Insert default configurations for each model
INSERT INTO settings (key, value, description, model_settings) 
VALUES (
    'model_config_qwen',
    'qwen/qwen-image-edit',
    'Configuration for Qwen Image Edit model',
    '{
        "go_fast": true,
        "aspect_ratio": "match_input_image",
        "output_format": "webp",
        "output_quality": 80
    }'::jsonb
) ON CONFLICT (key) DO NOTHING;

INSERT INTO settings (key, value, description, model_settings) 
VALUES (
    'model_config_flux_kontext_max',
    'black-forest-labs/flux-kontext-max',
    'Configuration for Flux Kontext Max model',
    '{
        "aspect_ratio": "match_input_image",
        "output_format": "png",
        "safety_tolerance": 2,
        "prompt_upsampling": false,
        "num_outputs": 1,
        "output_quality": 90
    }'::jsonb
) ON CONFLICT (key) DO NOTHING;

-- Similar for other models...
```

### 3. Configuration Repository

```go
// apps/worker/internal/settings/config_repository.go

type ConfigRepository interface {
    // GetModelConfig retrieves the configuration for a specific model
    GetModelConfig(ctx context.Context, modelID ModelID) (ModelConfig, error)
    
    // UpdateModelConfig updates the configuration for a specific model
    UpdateModelConfig(ctx context.Context, modelID ModelID, config ModelConfig, userID string) error
}
```

### 4. Updated Model Registry

```go
// apps/worker/internal/staging/model/registry.go

type ModelMetadata struct {
    ID            ModelID
    Name          string
    Description   string
    Version       string
    InputBuilder  ModelInputBuilder
    DefaultConfig ModelConfig      // NEW: Default configuration
    ConfigType    reflect.Type     // NEW: Type for config deserialization
}
```

### 5. Admin API Endpoints

```
GET  /api/v1/admin/models/:modelId/config          # Get model configuration
PUT  /api/v1/admin/models/:modelId/config          # Update model configuration
GET  /api/v1/admin/models/:modelId/config/schema   # Get configuration schema (for UI)
```

### 6. Frontend Admin UI

Dynamic form generation based on model configuration schema:

```tsx
// apps/web/app/admin/settings/model-config.tsx

interface ModelConfigProps {
  modelId: string;
  config: ModelConfig;
  schema: ConfigSchema;
}

export function ModelConfigForm({ modelId, config, schema }: ModelConfigProps) {
  // Dynamically generate form fields based on schema
  // - Text inputs for strings
  // - Number inputs with min/max
  // - Dropdowns for options
  // - Checkboxes for booleans
  // - Tooltips for descriptions
}
```

## Implementation Plan

### Phase 1: Core Infrastructure
1. Create configuration structs for each model
2. Add model_settings column to settings table
3. Update model registry to include config metadata
4. Create configuration repository

### Phase 2: Worker Integration
5. Update InputBuilders to accept ModelConfig
6. Modify staging service to load config from database
7. Add config validation
8. Update tests

### Phase 3: API Layer
9. Add configuration endpoints to admin handler
10. Add config schema generation
11. Update API documentation

### Phase 4: Frontend
12. Create model configuration UI component
13. Add form validation
14. Display current vs default configs
15. Add reset to defaults button

### Phase 5: Testing & Documentation
16. Integration tests for config management
17. Update documentation
18. Migration guide for existing deployments

## Benefits

1. **No code changes required** to tweak model parameters
2. **Per-model tuning** for optimal results
3. **A/B testing** different configurations
4. **Audit trail** of configuration changes
5. **Discoverable** - UI shows all available options
6. **Type-safe** - Struct validation ensures correctness

## Migration Strategy

1. Existing models continue to work with hardcoded defaults
2. Gradually migrate each model to use config system
3. Database migration adds configs for all existing models
4. Old code paths deprecated but not removed immediately
