# Model Registry Architecture

This document describes the model registry system used by the staging service to support multiple AI models with different API contracts.

## Overview

The staging service supports multiple Replicate AI models for virtual staging. Each model has its own API contract (input parameters), which is handled through a model registry system.

## Architecture

### Package Structure

The model registry is now organized in a dedicated package:
- **Location**: `apps/worker/internal/staging/model/`
- **Files**:
  - `registry.go` - Core registry and interface definitions
  - `qwen.go` - Qwen Image Edit model implementation
  - `flux_kontext.go` - Flux Kontext Max model implementation
  - `seedream.go` - Seedream model implementations (supports multiple versions)
  - `*_test.go` - Comprehensive test files (100% coverage)

### Components

1. **Model Enum**: A predefined enum of supported models
2. **ModelInputBuilder Interface**: Defines how to build input parameters for each model
3. **Model Registry**: Maps model IDs to their input builders and metadata
4. **DefaultService**: Uses the registry to build appropriate inputs for the selected model

### Model Enum

Models are defined as constants in the codebase (not configuration files). This ensures type safety and prevents runtime errors from invalid model names.

```go
type ModelID string

const (
    ModelQwenImageEdit  ModelID = "qwen/qwen-image-edit"
    ModelFluxKontextMax ModelID = "black-forest-labs/flux-kontext-max"
    ModelSeedream3      ModelID = "bytedance/seedream-3"
    ModelSeedream4      ModelID = "bytedance/seedream-4"
    ModelGPTImage1      ModelID = "openai/gpt-image-1"
    ModelGPTImage1_5    ModelID = "openai/gpt-image-1.5"
    // Additional models can be added here
)
```

### ModelInputBuilder Interface

Each model has a specific API contract. The `ModelInputBuilder` interface abstracts the creation of model-specific input parameters:

```go
type ModelInputBuilder interface {
    // BuildInput creates the input parameters for the model
    BuildInput(ctx context.Context, req *ModelInputRequest) (replicate.PredictionInput, error)
    
    // Validate checks if the request is valid for this model
    Validate(req *ModelInputRequest) error
}
```

### Model Metadata

Each registered model includes metadata:

- **ID**: Unique identifier (e.g., "qwen/qwen-image-edit")
- **Name**: Human-readable name
- **Description**: Brief description of the model's capabilities
- **Version**: Model version for tracking
- **InputBuilder**: Implementation of ModelInputBuilder for this model

## Supported Models

### 1. Qwen Image Edit

- **ID**: `qwen/qwen-image-edit`
- **Description**: Fast image editing model optimized for staging
- **Package Location**: `apps/worker/internal/staging/model/qwen.go`
- **Parameters**:
  - `image` (string, required): Base64-encoded image data URL
  - `prompt` (string, required): Editing instructions
  - `go_fast` (bool): Enable fast mode (default: true)
  - `aspect_ratio` (string): Output aspect ratio (default: "match_input_image")
  - `output_format` (string): Output format (default: "webp")
  - `output_quality` (int): Output quality 1-100 (default: 80)
  - `seed` (int, optional): Random seed for reproducibility

### 2. Flux Kontext Max

- **ID**: `black-forest-labs/flux-kontext-max`
- **Description**: High-quality image generation and editing with advanced context understanding
- **Package Location**: `apps/worker/internal/staging/model/flux_kontext.go`
- **Parameters**:
  - `prompt` (string, required): Text description or editing instruction
  - `input_image` (string, optional): Base64-encoded image data URL for image editing
  - `aspect_ratio` (string): Output aspect ratio (default: "match_input_image")
  - `output_format` (string): Output format - "jpg" or "png" (default: "png")
  - `safety_tolerance` (int): Safety level 0-6, 2 is max with input images (default: 2)
  - `prompt_upsampling` (bool): Automatic prompt improvement (default: false)
  - `seed` (int, optional): Random seed for reproducibility

### 3. Seedream Models

ByteDance's Seedream family of models provides unified text-to-image generation and precise editing capabilities. All versions share a common API structure and are implemented using a single flexible input builder.

**Package Location**: `apps/worker/internal/staging/model/seedream.go`

#### Seedream 3

- **ID**: `bytedance/seedream-3`
- **Description**: Unified text-to-image generation and precise editing
- **Cost**: $0.03 per output image
- **Parameters**:
  - `prompt` (string, required): Text description or editing instruction
  - `image_input` (array of strings, optional): Base64-encoded image data URLs for image editing
  - `size` (string): Image resolution (default: "2K")
  - `aspect_ratio` (string): Output aspect ratio (default: "match_input_image")
  - `enhance_prompt` (bool): Enable prompt enhancement for higher quality (default: true)
  - `max_images` (int): Maximum number of images to generate (default: 1)
  - `seed` (int, optional): Random seed for reproducibility

#### Seedream 4

- **ID**: `bytedance/seedream-4`
- **Description**: Latest version with support for up to 4K resolution
- **Cost**: $0.03 per output image
- **Parameters**:
  - `prompt` (string, required): Text description or editing instruction
  - `image_input` (array of strings, optional): Base64-encoded image data URLs for image editing (supports 1-10 images)
  - `size` (string): Image resolution - "1K" (1024px), "2K" (2048px), or "4K" (4096px) (default: "2K")
  - `aspect_ratio` (string): Output aspect ratio (default: "match_input_image")
  - `enhance_prompt` (bool): Enable prompt enhancement for higher quality (default: true)
  - `max_images` (int): Maximum number of images to generate (1-15) (default: 1)
  - `seed` (int, optional): Random seed for reproducibility
- **Features**:
  - Supports both text-to-image and image-to-image workflows
  - High-resolution output up to 4K (4096px)
  - Natural language editing commands
  - Batch and multi-reference support
  - Faster inference than previous generations

**Note**: The `SeedreamInputBuilder` is designed to support all Seedream versions with a common API structure, making it easy to add future Seedream models.

### Future Models

Additional models can be added by:

1. Defining a new constant in the ModelID enum
2. Creating a new implementation of ModelInputBuilder
3. Registering the model in the registry with its metadata
4. Adding documentation to this file

## Configuration

### Current Approach (Being Replaced)

Previously, the model version was stored in `config/shared.yml`:

```yaml
replicate:
  model_version: qwen/qwen-image-edit
```

### New Approach

Models are now defined in code as constants. Model selection will eventually be exposed through an admin-only configuration UI, allowing:

- Selection of the active model
- Configuration of model-specific parameters
- Per-project or global model settings

## Usage

### Initializing the Service

```go
import (
    "github.com/real-staging-ai/worker/internal/staging"
    "github.com/real-staging-ai/worker/internal/staging/model"
)

// Create service with specific model
stagingService, err := staging.NewDefaultService(ctx, &staging.ServiceConfig{
    ModelID:        model.ModelQwenImageEdit,     // or model.ModelFluxKontextMax
    BucketName:     cfg.S3Bucket(),
    ReplicateToken: cfg.Replicate.APIToken,
    // ... other config
})
```

### Staging an Image

The service automatically uses the correct input builder for the configured model:

```go
stagedURL, err := stagingService.StageImage(ctx, &staging.StagingRequest{
    ImageID:     "img-123",
    OriginalURL: "s3://bucket/uploads/original.jpg",
    RoomType:    ptr("living_room"),
    Style:       ptr("modern"),
    Seed:        ptr(int64(12345)),
})
```

## Testing

Each model input builder must have:

- Unit tests with 100% coverage
- Tests for all input parameters
- Validation error tests
- Integration tests with mock Replicate API

## Future Enhancements

1. **Admin UI**: Web interface for model selection and configuration
2. **Per-Project Models**: Allow different models per project
3. **A/B Testing**: Support running multiple models for comparison
4. **Cost Tracking**: Track API costs per model
5. **Performance Metrics**: Monitor quality and speed by model
6. **Model Fallback**: Automatic fallback if primary model fails

## Migration Notes

When migrating from the old system:

1. Remove `model_version` from config files
2. Update worker initialization to use `ModelID` constant
3. The default model remains `ModelQwenImageEdit`
4. No changes required to the database or API contracts
