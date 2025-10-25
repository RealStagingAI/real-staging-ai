package model

import (
	"context"
	"testing"
)

func TestNewSeedreamInputBuilder(t *testing.T) {
	t.Run("success: creates new builder", func(t *testing.T) {
		builder := NewSeedreamInputBuilder()
		if builder == nil {
			t.Fatal("expected builder to be non-nil")
		}
	})
}

func TestSeedreamInputBuilder_BuildInput(t *testing.T) {
	t.Run("success: builds input with image", func(t *testing.T) {
		builder := NewSeedreamInputBuilder()
		req := &ModelInputRequest{
			ImageDataURL: "data:image/png;base64,test",
			Prompt:       "Modern living room with minimalist furniture",
		}

		input, err := builder.BuildInput(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify required fields
		if input["prompt"] != req.Prompt {
			t.Errorf("expected prompt %q, got %q", req.Prompt, input["prompt"])
		}

		// Verify image_input is set as array
		imageInput, ok := input["image_input"].([]string)
		if !ok {
			t.Fatal("expected image_input to be []string")
		}
		if len(imageInput) != 1 {
			t.Fatalf("expected image_input to have 1 element, got %d", len(imageInput))
		}
		if imageInput[0] != req.ImageDataURL {
			t.Errorf("expected image_input[0] to be %q, got %q", req.ImageDataURL, imageInput[0])
		}

		// Verify default parameters (from SeedreamConfig defaults)
		if input["aspect_ratio"] != "1:1" {
			t.Errorf("expected aspect_ratio to be '1:1', got %q", input["aspect_ratio"])
		}
		if input["num_inference_steps"] != 50 {
			t.Errorf("expected num_inference_steps to be 50, got %v", input["num_inference_steps"])
		}
		if input["guidance_scale"] != 7.5 {
			t.Errorf("expected guidance_scale to be 7.5, got %v", input["guidance_scale"])
		}
		if input["output_quality"] != 95 {
			t.Errorf("expected output_quality to be 95, got %v", input["output_quality"])
		}
	})

	t.Run("success: builds input without image", func(t *testing.T) {
		builder := NewSeedreamInputBuilder()
		req := &ModelInputRequest{
			Prompt: "Modern living room with minimalist furniture",
		}

		input, err := builder.BuildInput(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify image_input is not set
		if _, exists := input["image_input"]; exists {
			t.Error("expected image_input to not be set when ImageDataURL is empty")
		}

		// Verify prompt is still set
		if input["prompt"] != req.Prompt {
			t.Errorf("expected prompt %q, got %q", req.Prompt, input["prompt"])
		}
	})

	t.Run("success: builds input with seed", func(t *testing.T) {
		builder := NewSeedreamInputBuilder()
		seed := int64(42)
		req := &ModelInputRequest{
			ImageDataURL: "data:image/png;base64,test",
			Prompt:       "Modern living room",
			Seed:         &seed,
		}

		input, err := builder.BuildInput(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if input["seed"] != seed {
			t.Errorf("expected seed to be %d, got %v", seed, input["seed"])
		}
	})

	t.Run("success: builds input with all parameters", func(t *testing.T) {
		builder := NewSeedreamInputBuilder()
		seed := int64(12345)
		req := &ModelInputRequest{
			ImageDataURL: "data:image/png;base64,testimage",
			Prompt:       "Cozy bedroom with rustic furniture",
			Seed:         &seed,
		}

		input, err := builder.BuildInput(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify all fields are set correctly
		if input["prompt"] != req.Prompt {
			t.Errorf("expected prompt %q, got %q", req.Prompt, input["prompt"])
		}
		imageInput := input["image_input"].([]string)
		if imageInput[0] != req.ImageDataURL {
			t.Errorf("expected image_input %q, got %q", req.ImageDataURL, imageInput[0])
		}
		if input["seed"] != seed {
			t.Errorf("expected seed %d, got %v", seed, input["seed"])
		}
		// Verify config-based parameters
		if input["aspect_ratio"] == nil {
			t.Error("expected aspect_ratio to be set")
		}
		if input["num_inference_steps"] == nil {
			t.Error("expected num_inference_steps to be set")
		}
	})

	t.Run("fail: nil request", func(t *testing.T) {
		builder := NewSeedreamInputBuilder()

		_, err := builder.BuildInput(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error for nil request")
		}

		expectedMsg := "request cannot be nil"
		if err.Error() != expectedMsg {
			t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
		}
	})

	t.Run("fail: empty prompt", func(t *testing.T) {
		builder := NewSeedreamInputBuilder()
		req := &ModelInputRequest{
			ImageDataURL: "data:image/png;base64,test",
			Prompt:       "",
		}

		_, err := builder.BuildInput(context.Background(), req)
		if err == nil {
			t.Fatal("expected error for empty prompt")
		}

		expectedMsg := "prompt is required"
		if err.Error() != expectedMsg {
			t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
		}
	})
}

func TestSeedreamInputBuilder_Validate(t *testing.T) {
	t.Run("success: valid request with prompt only", func(t *testing.T) {
		builder := NewSeedreamInputBuilder()
		req := &ModelInputRequest{
			Prompt: "Modern kitchen design",
		}

		err := builder.Validate(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("success: valid request with image and prompt", func(t *testing.T) {
		builder := NewSeedreamInputBuilder()
		req := &ModelInputRequest{
			ImageDataURL: "data:image/png;base64,test",
			Prompt:       "Add modern furniture",
		}

		err := builder.Validate(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("success: valid request with all fields", func(t *testing.T) {
		builder := NewSeedreamInputBuilder()
		seed := int64(999)
		req := &ModelInputRequest{
			ImageDataURL: "data:image/png;base64,test",
			Prompt:       "Scandinavian style living room",
			Seed:         &seed,
		}

		err := builder.Validate(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("fail: nil request", func(t *testing.T) {
		builder := NewSeedreamInputBuilder()

		err := builder.Validate(nil)
		if err == nil {
			t.Fatal("expected error for nil request")
		}

		expectedMsg := "request cannot be nil"
		if err.Error() != expectedMsg {
			t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
		}
	})

	t.Run("fail: empty prompt", func(t *testing.T) {
		builder := NewSeedreamInputBuilder()
		req := &ModelInputRequest{
			ImageDataURL: "data:image/png;base64,test",
			Prompt:       "",
		}

		err := builder.Validate(req)
		if err == nil {
			t.Fatal("expected error for empty prompt")
		}

		expectedMsg := "prompt is required"
		if err.Error() != expectedMsg {
			t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
		}
	})
}
