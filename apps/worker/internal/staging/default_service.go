package staging

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/replicate/replicate-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/real-staging-ai/worker/internal/logging"
	"github.com/real-staging-ai/worker/internal/staging/model"
	"github.com/real-staging-ai/worker/internal/staging/prompt"
)

// DefaultService implements the Service interface using Replicate AI and S3.
type DefaultService struct {
	s3Client        *s3.Client
	bucketName      string
	replicateClient *replicate.Client
	modelID         model.ModelID
	registry        *model.ModelRegistry
	promptLib       *prompt.Library
	configRepo      ConfigRepository // For loading model configurations
}

// Ensure DefaultService implements Service interface.
var _ Service = (*DefaultService)(nil)

// awsConfigLoader allows overriding AWS config loading in tests.
var awsConfigLoader = config.LoadDefaultConfig

// ServiceConfig holds configuration for the staging service.
type ServiceConfig struct {
	BucketName     string
	ReplicateToken string
	ModelID        model.ModelID
	S3Endpoint     string
	S3Region       string
	S3AccessKey    string
	S3SecretKey    string
	S3UsePathStyle bool
	AppEnv         string
	ConfigRepo     ConfigRepository // Optional: for loading model configs from database
}

// NewDefaultService creates a new DefaultService instance using provided configuration.
func NewDefaultService(ctx context.Context, cfg *ServiceConfig) (*DefaultService, error) {
	// Validate required configuration
	if cfg.BucketName == "" {
		return nil, fmt.Errorf("bucket name is required")
	}
	if cfg.ReplicateToken == "" {
		return nil, fmt.Errorf("replicate API token is required")
	}

	// Use default model if not specified
	modelID := cfg.ModelID
	if modelID == "" {
		modelID = model.ModelQwenImageEdit
	}

	// Initialize model registry
	registry := model.NewModelRegistry()

	// Validate model exists
	if !registry.Exists(modelID) {
		return nil, fmt.Errorf("unsupported model: %s", modelID)
	}

	bucketName := cfg.BucketName
	replicateToken := cfg.ReplicateToken

	// Create Replicate client
	replicateClient, err := replicate.NewClient(replicate.WithToken(replicateToken))
	if err != nil {
		return nil, fmt.Errorf("failed to create Replicate client: %w", err)
	}

	// Initialize S3 client
	var awsCfg aws.Config

	if cfg.AppEnv == "test" {
		awsCfg, err = awsConfigLoader(ctx,
			config.WithRegion("us-east-1"),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "test")),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config for test: %w", err)
		}

		s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String("http://localhost:4566")
			o.UsePathStyle = true
		})

		return &DefaultService{
			s3Client:        s3Client,
			bucketName:      bucketName,
			replicateClient: replicateClient,
			modelID:         modelID,
			registry:        registry,
			promptLib:       prompt.New(),
			configRepo:      cfg.ConfigRepo,
		}, nil
	}

	// If a custom S3 endpoint is provided (e.g., MinIO), configure client for dev/local
	if cfg.S3Endpoint != "" {
		region := cfg.S3Region
		if region == "" {
			region = "us-west-1"
		}
		awsCfg, err = awsConfigLoader(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.S3AccessKey, cfg.S3SecretKey, "")),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config for dev: %w", err)
		}

		s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.S3Endpoint)
			if cfg.S3UsePathStyle {
				o.UsePathStyle = true
			}
		})

		return &DefaultService{
			s3Client:        s3Client,
			bucketName:      bucketName,
			replicateClient: replicateClient,
			modelID:         modelID,
			registry:        registry,
			promptLib:       prompt.New(),
			configRepo:      cfg.ConfigRepo,
		}, nil
	}

	// Use default AWS config for production
	awsCfg, err = awsConfigLoader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg)

	return &DefaultService{
		s3Client:        s3Client,
		bucketName:      bucketName,
		replicateClient: replicateClient,
		modelID:         modelID,
		registry:        registry,
		promptLib:       prompt.New(),
		configRepo:      cfg.ConfigRepo,
	}, nil
}

// StageImage processes an image with AI staging and returns the staged image URL in S3.
func (s *DefaultService) StageImage(ctx context.Context, req *StagingRequest) (string, error) {
	log := logging.Default()
	tracer := otel.Tracer("real-staging-worker/staging")
	ctx, span := tracer.Start(ctx, "staging.StageImage")
	span.SetAttributes(
		attribute.String("image.id", req.ImageID),
		attribute.String("image.original_url", req.OriginalURL),
	)
	defer span.End()

	// Extract the S3 file key from the original URL
	fileKey, err := extractS3KeyFromURL(req.OriginalURL)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid S3 URL")
		return "", fmt.Errorf("failed to extract S3 key from URL: %w", err)
	}

	// Download the original image from S3
	originalImage, err := s.DownloadFromS3(ctx, fileKey)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "S3 download failed")
		return "", fmt.Errorf("failed to download original image: %w", err)
	}
	defer func() {
		if err := originalImage.Close(); err != nil {
			span.RecordError(err)
			log.Error(ctx, "failed to close original image", "error", err)
		}
	}()

	// Read the image content
	imageBytes, err := io.ReadAll(originalImage)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "read image failed")
		return "", fmt.Errorf("failed to read image content: %w", err)
	}

	// Convert to base64 data URL for Replicate
	mimeType := http.DetectContentType(imageBytes)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(imageBytes))

	// Build the prompt using library or custom prompt
	promptText := s.buildPrompt(req.RoomType, req.Style, req.Prompt)

	// Call Replicate AI to stage the image
	stagedImageURL, err := s.callReplicateAPI(ctx, dataURL, promptText, req.Seed)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Replicate API failed")
		return "", fmt.Errorf("failed to stage image with Replicate: %w", err)
	}

	// Download the staged image from Replicate's CDN
	stagedImageBytes, err := s.downloadFromURL(ctx, stagedImageURL)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "download staged image failed")
		return "", fmt.Errorf("failed to download staged image: %w", err)
	}

	// Upload the staged image to S3
	stagedURL, err := s.UploadToS3(ctx, req.ImageID, bytes.NewReader(stagedImageBytes), "image/jpeg")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "S3 upload failed")
		return "", fmt.Errorf("failed to upload staged image: %w", err)
	}

	span.SetStatus(codes.Ok, "staging completed")
	return stagedURL, nil
}

// DownloadFromS3 downloads a file from S3 and returns its content.
func (s *DefaultService) DownloadFromS3(ctx context.Context, fileKey string) (io.ReadCloser, error) {
	tracer := otel.Tracer("real-staging-worker/staging")
	_, span := tracer.Start(ctx, "staging.DownloadFromS3")
	span.SetAttributes(attribute.String("s3.key", fileKey))
	defer span.End()

	result, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "GetObject failed")
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}

	span.SetStatus(codes.Ok, "download completed")
	return result.Body, nil
}

// UploadToS3 uploads a file to S3 and returns the public URL.
func (s *DefaultService) UploadToS3(
	ctx context.Context, imageID string, content io.Reader, contentType string,
) (string, error) {
	tracer := otel.Tracer("real-staging-worker/staging")
	_, span := tracer.Start(ctx, "staging.UploadToS3")
	span.SetAttributes(attribute.String("image.id", imageID))
	defer span.End()

	// Generate the S3 key for the staged image
	fileKey := fmt.Sprintf("staged/%s/%s-staged.jpg", imageID[:8], imageID)

	// Upload to S3
	// Set Cache-Control for Render Edge Caching: staged images are immutable, cache for 1 year
	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(s.bucketName),
		Key:          aws.String(fileKey),
		Body:         content,
		ContentType:  aws.String(contentType),
		CacheControl: aws.String("public, max-age=31536000, immutable"),
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "PutObject failed")
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Construct the public URL
	// In production, this would be the S3 URL or CloudFront URL
	// For now, we'll return the key which can be used with presigned URLs
	publicURL := fmt.Sprintf("s3://%s/%s", s.bucketName, fileKey)

	span.SetStatus(codes.Ok, "upload completed")
	return publicURL, nil
}

// callReplicateAPI calls the Replicate API to stage an image.
func (s *DefaultService) callReplicateAPI(
	ctx context.Context, imageDataURL, prompt string, seed *int64,
) (string, error) {
	tracer := otel.Tracer("real-staging-worker/staging")
	ctx, span := tracer.Start(ctx, "staging.callReplicateAPI")
	span.SetAttributes(
		attribute.String("model", string(s.modelID)),
		attribute.String("prompt", prompt),
	)
	defer span.End()

	// Get the model metadata from registry
	modelMeta, err := s.registry.Get(s.modelID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "model not found")
		return "", fmt.Errorf("failed to get model metadata: %w", err)
	}

	// Load model configuration from database (optional)
	var modelConfig model.ModelConfig
	if s.configRepo != nil {
		modelConfig, err = s.configRepo.GetModelConfig(ctx, s.modelID)
		if err != nil {
			// Log warning but continue with defaults
			log := logging.Default()
			log.Warn(ctx, "failed to load model config, using defaults", "error", err, "model", s.modelID)
			modelConfig = nil
		}
	}

	// Build the input parameters using the model's input builder
	inputReq := &model.ModelInputRequest{
		ImageDataURL: imageDataURL,
		Prompt:       prompt,
		Seed:         seed,
		Config:       modelConfig, // Will use defaults if nil
	}

	input, err := modelMeta.InputBuilder.BuildInput(ctx, inputReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "input build failed")
		return "", fmt.Errorf("failed to build model input: %w", err)
	}

	// Create and run the prediction
	webhook := replicate.Webhook{
		URL:    "", // No webhook for now
		Events: []replicate.WebhookEventType{},
	}

	prediction, err := s.replicateClient.CreatePrediction(ctx, string(s.modelID), input, &webhook, false)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "CreatePrediction failed")
		return "", fmt.Errorf("failed to create prediction: %w", err)
	}

	// Wait for the prediction to complete (with timeout)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(5 * time.Minute)

	for {
		select {
		case <-timeout:
			err := fmt.Errorf("prediction timed out after 5 minutes")
			span.RecordError(err)
			span.SetStatus(codes.Error, "prediction timeout")
			return "", err

		case <-ticker.C:
			pred, err := s.replicateClient.GetPrediction(ctx, prediction.ID)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "GetPrediction failed")
				return "", fmt.Errorf("failed to get prediction status: %w", err)
			}

			switch pred.Status {
			case replicate.Succeeded:
				// Extract the output URL from the prediction
				if pred.Output == nil {
					err := fmt.Errorf("prediction succeeded but output is nil")
					span.RecordError(err)
					span.SetStatus(codes.Error, "nil output")
					return "", err
				}

				// The output can be a string URL or an array of URLs
				var outputURL string
				switch v := pred.Output.(type) {
				case string:
					outputURL = v
				case []interface{}:
					if len(v) > 0 {
						if url, ok := v[0].(string); ok {
							outputURL = url
						}
					}
				}

				if outputURL == "" {
					err := fmt.Errorf("could not extract output URL from prediction")
					span.RecordError(err)
					span.SetStatus(codes.Error, "invalid output format")
					return "", err
				}

				span.SetStatus(codes.Ok, "prediction succeeded")
				return outputURL, nil

			case replicate.Failed:
				err := fmt.Errorf("prediction failed: %v", pred.Error)
				span.RecordError(err)
				span.SetStatus(codes.Error, "prediction failed")
				return "", err

			case replicate.Canceled:
				err := fmt.Errorf("prediction was canceled")
				span.RecordError(err)
				span.SetStatus(codes.Error, "prediction canceled")
				return "", err

			case replicate.Processing, replicate.Starting:
				// Continue polling
				continue

			default:
				err := fmt.Errorf("unknown prediction status: %s", pred.Status)
				span.RecordError(err)
				span.SetStatus(codes.Error, "unknown status")
				return "", err
			}
		}
	}
}

// buildPrompt constructs the AI prompt using the library or custom prompt.
// If customPrompt is provided, it takes precedence.
// Otherwise, retrieves the appropriate prompt from the library based on room type and style.
func (s *DefaultService) buildPrompt(roomType, style, customPrompt *string) string {
	// Extract values from pointers, using empty strings as defaults
	roomTypeStr := ""
	if roomType != nil {
		roomTypeStr = *roomType
	}

	styleStr := ""
	if style != nil {
		styleStr = *style
	}

	customPromptStr := ""
	if customPrompt != nil {
		customPromptStr = *customPrompt
	}

	// Use the prompt library to build the final prompt
	return s.promptLib.Build(roomTypeStr, styleStr, customPromptStr)
}

// downloadFromURL downloads content from an HTTP(S) URL.
func (s *DefaultService) downloadFromURL(ctx context.Context, url string) ([]byte, error) {
	log := logging.Default()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download from URL: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// extractS3KeyFromURL extracts the S3 key from a URL.
// Handles formats like:
//   - http://localhost:9000/bucket-name/uploads/...
//   - https://bucket-name.s3.amazonaws.com/uploads/...
//   - s3://bucket-name/uploads/...
func extractS3KeyFromURL(rawURL string) (string, error) {
	// Handle s3:// URLs
	if strings.HasPrefix(rawURL, "s3://") {
		parts := strings.SplitN(strings.TrimPrefix(rawURL, "s3://"), "/", 2)
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid s3:// URL format")
		}
		return parts[1], nil
	}

	// Parse HTTP(S) URLs
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Remove leading slash and bucket name if path-style
	path := strings.TrimPrefix(u.Path, "/")

	// If the first path segment is the bucket name (path-style), remove it
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 2 {
		// Check if first part looks like a bucket name
		if strings.Contains(parts[0], "real-staging") {
			return parts[1], nil
		}
	}

	return path, nil
}
