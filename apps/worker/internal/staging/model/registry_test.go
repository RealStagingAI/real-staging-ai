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
		if len(models) != 6 {
			t.Errorf("expected 6 models to be registered, got %d", len(models))
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
	t.Run("success: retrieves Qwen model", func(t *testing.T) {
		registry := NewModelRegistry()

		model, err := registry.Get(ModelQwenImageEdit)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if model.ID != ModelQwenImageEdit {
			t.Errorf("expected model ID to be %s, got %s", ModelQwenImageEdit, model.ID)
		}

		if model.Name == "" {
			t.Error("expected model to have a name")
		}

		if model.InputBuilder == nil {
			t.Error("expected model to have an input builder")
		}
	})

	t.Run("success: retrieves Flux Kontext Max model", func(t *testing.T) {
		registry := NewModelRegistry()

		model, err := registry.Get(ModelFluxKontextMax)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if model.ID != ModelFluxKontextMax {
			t.Errorf("expected model ID to be %s, got %s", ModelFluxKontextMax, model.ID)
		}

		if model.Name != "Flux Kontext Max" {
			t.Errorf("expected model name to be 'Flux Kontext Max', got %s", model.Name)
		}

		if model.InputBuilder == nil {
			t.Error("expected model to have an input builder")
		}
	})

	t.Run("success: retrieves Flux Kontext Pro model", func(t *testing.T) {
		registry := NewModelRegistry()

		model, err := registry.Get(ModelFluxKontextPro)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if model.ID != ModelFluxKontextPro {
			t.Errorf("expected model ID to be %s, got %s", ModelFluxKontextPro, model.ID)
		}

		if model.Name != "Flux Kontext Pro" {
			t.Errorf("expected model name to be 'Flux Kontext Pro', got %s", model.Name)
		}

		if model.InputBuilder == nil {
			t.Error("expected model to have an input builder")
		}
	})

	t.Run("success: retrieves Seedream-3 model", func(t *testing.T) {
		registry := NewModelRegistry()

		model, err := registry.Get(ModelSeedream3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if model.ID != ModelSeedream3 {
			t.Errorf("expected model ID to be %s, got %s", ModelSeedream3, model.ID)
		}

		if model.Name != "Seedream 3" {
			t.Errorf("expected model name to be 'Seedream 3', got %s", model.Name)
		}

		if model.InputBuilder == nil {
			t.Error("expected model to have an input builder")
		}
	})

	t.Run("success: retrieves Seedream-4 model", func(t *testing.T) {
		registry := NewModelRegistry()

		model, err := registry.Get(ModelSeedream4)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if model.ID != ModelSeedream4 {
			t.Errorf("expected model ID to be %s, got %s", ModelSeedream4, model.ID)
		}

		if model.Name != "Seedream 4" {
			t.Errorf("expected model name to be 'Seedream 4', got %s", model.Name)
		}

		if model.InputBuilder == nil {
			t.Error("expected model to have an input builder")
		}
	})

	t.Run("success: retrieves GPT Image 1 model", func(t *testing.T) {
		registry := NewModelRegistry()

		model, err := registry.Get(ModelGPTImage1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if model.ID != ModelGPTImage1 {
			t.Errorf("expected model ID to be %s, got %s", ModelGPTImage1, model.ID)
		}

		if model.Name != "GPT Image 1" {
			t.Errorf("expected model name to be 'GPT Image 1', got %s", model.Name)
		}

		if model.InputBuilder == nil {
			t.Error("expected model to have an input builder")
		}
	})

	t.Run("fail: model not found", func(t *testing.T) {
		registry := NewModelRegistry()

		_, err := registry.Get(ID("nonexistent/model"))
		if err == nil {
			t.Fatal("expected error for nonexistent model")
		}

		expectedMsg := "model not found: nonexistent/model"
		if err.Error() != expectedMsg {
			t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
		}
	})
}

func TestModelRegistry_List(t *testing.T) {
	t.Run("success: lists all registered models", func(t *testing.T) {
		registry := NewModelRegistry()

		models := registry.List()

		if len(models) != 6 {
			t.Errorf("expected 6 models to be registered, got %d", len(models))
		}

		// Verify all models are in the list
		foundQwen := false
		foundFluxMax := false
		foundFluxPro := false
		foundSeedream3 := false
		foundSeedream4 := false
		foundGPTImage1 := false
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
