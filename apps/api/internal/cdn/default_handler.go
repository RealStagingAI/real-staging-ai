package cdn

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/real-staging-ai/api/internal/image"
	"github.com/real-staging-ai/api/internal/logging"
)

// DefaultHandler provides Echo HTTP handlers for CDN proxy operations.
type DefaultHandler struct {
	log          logging.Logger
	cdnURL       string
	imageService image.Service
}

// NewDefaultHandler constructs a CDN handler with the provided CDN URL and image service.
func NewDefaultHandler(log logging.Logger, cdnURL string, imageService image.Service) *DefaultHandler {
	return &DefaultHandler{
		log:          log,
		cdnURL:       cdnURL,
		imageService: imageService,
	}
}

// Ensure DefaultHandler implements Handler.
var _ Handler = (*DefaultHandler)(nil)

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// ProxyImage handles GET /api/v1/images/:id/cdn/:kind
// This proxies image requests to the Cloudflare CDN Worker, adding the required
// Authorization header that Next.js Image component can't provide directly.
//
// Path params:
// - id: UUID of the image
// - kind: original|staged
func (h *DefaultHandler) ProxyImage(c echo.Context) error {
	imageID := c.Param("id")
	kind := strings.ToLower(strings.TrimSpace(c.Param("kind")))

	ctx := c.Request().Context()
	h.log.Info(ctx, "CDN ProxyImage called", "imageID", imageID, "kind", kind)
	
	// Debug: Check if JWT token is in context
	user := c.Get("user")
	h.log.Info(ctx, "CDN ProxyImage: checking user in context", "userType", fmt.Sprintf("%T", user), "userNil", user == nil)

	// Validate image ID
	if imageID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "bad_request", Message: "image id is required"})
	}
	if _, err := uuid.Parse(imageID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "bad_request", Message: "invalid image id format"})
	}

	// Validate kind parameter
	if kind != "original" && kind != "staged" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "invalid kind parameter. Must be 'original' or 'staged'",
		})
	}
	
	h.log.Info(ctx, "CDN ProxyImage: validation passed, fetching image", "imageID", imageID)

	// Verify user owns the image
	img, err := h.imageService.GetImageByID(ctx, imageID)
	if err != nil {
		h.log.Warn(ctx, "CDN ProxyImage: image not found", "imageID", imageID, "error", err)
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "image not found"})
	}

	// Verify image has the requested variant
	if kind == "staged" && (img.StagedURL == nil || *img.StagedURL == "") {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "image has no staged variant",
		})
	}

	// Check if CDN is configured
	if h.cdnURL == "" {
		return c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "service_unavailable",
			Message: "CDN not configured",
		})
	}

	// Extract JWT token from Echo context (populated by auth middleware)
	// This allows <img> tags to load images without sending Authorization headers
	token, ok := c.Get("user").(*jwt.Token)
	if !ok {
		h.log.Error(ctx, "CDN ProxyImage: failed to get JWT token from context", "imageID", imageID)
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid or missing JWT token",
		})
	}

	h.log.Debug(ctx, "CDN ProxyImage: successfully extracted JWT token from context", "imageID", imageID)

	// Reconstruct Authorization header for forwarding to CDN Worker
	authHeader := fmt.Sprintf("Bearer %s", token.Raw)

	// Construct CDN request URL
	cdnRequestURL := fmt.Sprintf("%s/images/%s/%s", h.cdnURL, imageID, kind)

	// Create request to CDN
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cdnRequestURL, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "failed to create CDN request",
		})
	}

	// Forward the Authorization header to CDN
	req.Header.Set("Authorization", authHeader)

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.JSON(http.StatusBadGateway, ErrorResponse{
			Error:   "bad_gateway",
			Message: fmt.Sprintf("CDN request failed: %v", err),
		})
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			h.log.Error(ctx, "Error closing CDN response body", "error", closeErr)
		}
	}()

	// Check CDN response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return c.JSON(resp.StatusCode, ErrorResponse{
			Error:   "cdn_error",
			Message: fmt.Sprintf("CDN returned %d: %s", resp.StatusCode, string(body)),
		})
	}

	// Get content type from CDN response
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	// Get cache status if available
	cacheStatus := resp.Header.Get("X-Cache-Status")

	// Set response headers
	c.Response().Header().Set("Content-Type", contentType)
	c.Response().Header().Set("Cache-Control", "private, max-age=3600") // Cache for 1 hour
	if cacheStatus != "" {
		c.Response().Header().Set("X-CDN-Cache-Status", cacheStatus)
	}

	// Stream the image data
	c.Response().WriteHeader(http.StatusOK)
	_, err = io.Copy(c.Response().Writer, resp.Body)
	if err != nil {
		// Can't return JSON error here as headers are already sent
		h.log.Error(ctx, "Error streaming CDN response", "error", err)
	}

	return nil
}
