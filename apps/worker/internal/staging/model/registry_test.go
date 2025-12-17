package model

import (
	"testing"
)

func TestNewModelRegistry(t *testing.T) {
	t.Run("success: creates registry with default models", func(t *testing.T) {
		registry := NewModelRegistry()

		if registry == nil {
			t.Fatal("expected registry to be non-nil")
		}

		// Verify Qwen model is registered
		if !registry.Exists(ModelQwenImageEdit) {
			t.Error("expected Qwen model to be registered")
		}

		// Verify Flux Kontext models are registered
		if !registry.Exists(ModelFluxKontextMax) {
			t.Error("expected Flux Kontext Max model to be registered")
		}
		if !registry.Exists(ModelFluxKontextPro) {
			t.Error("expected Flux Kontext Pro model to be registered")
		}

		// Verify Seedream models are registered
		if !registry.Exists(ModelSeedream3) {
			t.Error("expected Seedream-3 model to be registered")
		}
		if !registry.Exists(ModelSeedream4) {
			t.Error("expected Seedream-4 model to be registered")
		}
	})

	t.Run("success: registry has correct model count", func(t *testing.T) {
		registry := NewModelRegistry()

		models := registry.List()
		if len(models) != 7 {
			t.Errorf("expected 7 models to be registered, got %d", len(models))
		}
	})
}

func TestModelRegistry_Register(t *testing.T) {
	t.Run("success: registers a new model", func(t *testing.T) {
		registry := &ModelRegistry{
			models: make(map[ID]*ModelMetadata),
		}

		metadata := &ModelMetadata{
			ID:           ID("test/model"),
			Name:         "Test Model",
			Description:  "A test model",
			Version:      "v1",
			InputBuilder: NewQwenInputBuilder(),
		}

		registry.Register(metadata)

		if !registry.Exists(ID("test/model")) {
			t.Error("expected model to be registered")
		}
	})

	t.Run("success: overwrites existing model", func(t *testing.T) {
		registry := &ModelRegistry{
			models: make(map[ID]*ModelMetadata),
		}

		metadata1 := &ModelMetadata{
			ID:          ID("test/model"),
			Name:        "Test Model v1",
			Description: "Version 1",
		}

		metadata2 := &ModelMetadata{
			ID:          ID("test/model"),
			Name:        "Test Model v2",
			Description: "Version 2",
		}

		registry.Register(metadata1)
		registry.Register(metadata2)

		model, err := registry.Get(ID("test/model"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if model.Name != "Test Model v2" {
			t.Errorf("expected name to be 'Test Model v2', got %s", model.Name)
		}
	})
}

func TestModelRegistry_Get(t *testing.T) {
	tests := []struct {
		name           string
		modelID        ID
		wantName       string
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:        "success: retrieves Qwen model",
			modelID:     ModelQwenImageEdit,
			wantName:    "", // any name is fine for Qwen
			expectError: false,
		},
		{
			name:        "success: retrieves Flux Kontext Max model",
			modelID:     ModelFluxKontextMax,
			wantName:    "Flux Kontext Max",
			expectError: false,
		},
		{
			name:        "success: retrieves Flux Kontext Pro model",
			modelID:     ModelFluxKontextPro,
			wantName:    "Flux Kontext Pro",
			expectError: false,
		},
		{
			name:        "success: retrieves Seedream-3 model",
			modelID:     ModelSeedream3,
			wantName:    "Seedream 3",
			expectError: false,
		},
		{
			name:        "success: retrieves Seedream-4 model",
			modelID:     ModelSeedream4,
			wantName:    "Seedream 4",
			expectError: false,
		},
		{
			name:        "success: retrieves GPT Image 1 model",
			modelID:     ModelGPTImage1,
			wantName:    "GPT Image 1",
			expectError: false,
		},
		{
			name:        "success: retrieves GPT Image 1.5 model",
			modelID:     ModelGPTImage1_5,
			wantName:    "GPT Image 1.5",
			expectError: false,
		},
		{
			name:           "fail: model not found",
			modelID:        ID("nonexistent/model"),
			expectError:    true,
			expectedErrMsg: "model not found: nonexistent/model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewModelRegistry()

			model, err := registry.Get(tt.modelID)
			if tt.expectError {
				if err == nil {
					t.Fatal("expected error for nonexistent model")
				}
				if err.Error() != tt.expectedErrMsg {
					t.Errorf("expected error message %q, got %q", tt.expectedErrMsg, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if model.ID != tt.modelID {
				t.Errorf("expected model ID to be %s, got %s", tt.modelID, model.ID)
			}
			if tt.wantName != "" && model.Name != tt.wantName {
				t.Errorf("expected model name to be '%s', got %s", tt.wantName, model.Name)
			}
			if model.Name == "" {
				t.Error("expected model to have a name")
			}
			if model.InputBuilder == nil {
				t.Error("expected model to have an input builder")
			}
		})
	}
}

func TestModelRegistry_List(t *testing.T) {
	t.Run("success: lists all registered models", func(t *testing.T) {
		registry := NewModelRegistry()

		models := registry.List()

		if len(models) != 7 {
			t.Errorf("expected 7 models to be registered, got %d", len(models))
		}

		// Verify all models are in the list
		foundQwen := false
		foundFluxMax := false
		foundFluxPro := false
		foundSeedream3 := false
		foundSeedream4 := false
		foundGPTImage1 := false
		foundGPTImage1_5 := false
		for _, model := range models {
			if model.ID == ModelQwenImageEdit {
				foundQwen = true
			}
			if model.ID == ModelFluxKontextMax {
				foundFluxMax = true
			}
			if model.ID == ModelFluxKontextPro {
				foundFluxPro = true
			}
			if model.ID == ModelSeedream3 {
				foundSeedream3 = true
			}
			if model.ID == ModelSeedream4 {
				foundSeedream4 = true
			}
			if model.ID == ModelGPTImage1 {
				foundGPTImage1 = true
			}
			if model.ID == ModelGPTImage1_5 {
				foundGPTImage1_5 = true
			}
		}

		if !foundQwen {
			t.Error("expected Qwen model to be in the list")
		}
		if !foundFluxMax {
			t.Error("expected Flux Kontext Max model to be in the list")
		}
		if !foundFluxPro {
			t.Error("expected Flux Kontext Pro model to be in the list")
		}
		if !foundSeedream3 {
			t.Error("expected Seedream-3 model to be in the list")
		}
		if !foundSeedream4 {
			t.Error("expected Seedream-4 model to be in the list")
		}
		if !foundGPTImage1 {
			t.Error("expected GPT Image 1 model to be in the list")
		}
		if !foundGPTImage1_5 {
			t.Error("expected GPT Image 1.5 model to be in the list")
		}
	})

	t.Run("success: returns empty list for empty registry", func(t *testing.T) {
		registry := &ModelRegistry{
			models: make(map[ID]*ModelMetadata),
		}

		models := registry.List()

		if len(models) != 0 {
			t.Errorf("expected empty list, got %d models", len(models))
		}
	})
}

func TestModelRegistry_Exists(t *testing.T) {
	t.Run("success: returns true for Qwen model", func(t *testing.T) {
		registry := NewModelRegistry()

		if !registry.Exists(ModelQwenImageEdit) {
			t.Error("expected Qwen model to exist")
		}
	})

	t.Run("success: returns true for Flux Kontext Max model", func(t *testing.T) {
		registry := NewModelRegistry()

		if !registry.Exists(ModelFluxKontextMax) {
			t.Error("expected Flux Kontext Max model to exist")
		}
	})

	t.Run("success: returns true for Flux Kontext Pro model", func(t *testing.T) {
		registry := NewModelRegistry()

		if !registry.Exists(ModelFluxKontextPro) {
			t.Error("expected Flux Kontext Pro model to exist")
		}
	})

	t.Run("success: returns true for Seedream-3 model", func(t *testing.T) {
		registry := NewModelRegistry()

		if !registry.Exists(ModelSeedream3) {
			t.Error("expected Seedream-3 model to exist")
		}
	})

	t.Run("success: returns true for Seedream-4 model", func(t *testing.T) {
		registry := NewModelRegistry()

		if !registry.Exists(ModelSeedream4) {
			t.Error("expected Seedream-4 model to exist")
		}
	})

	t.Run("success: returns false for nonexistent model", func(t *testing.T) {
		registry := NewModelRegistry()

		if registry.Exists(ID("nonexistent/model")) {
			t.Error("expected nonexistent model to not exist")
		}
	})
}
