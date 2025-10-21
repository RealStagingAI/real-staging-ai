package http

import (
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ImageOwnershipResponse represents the response for image ownership checks
type ImageOwnershipResponse struct {
	ImageID    string `json:"image_id"`
	OwnerID    string `json:"owner_id"`
	HasAccess  bool   `json:"has_access"`
	S3Key      string `json:"s3_key,omitempty"`
}

// getImageOwnerHandler handles GET /v1/images/:id/owner
// Internal endpoint for Cloudflare Worker to verify image ownership
// and retrieve S3 key for authorized access
func (s *Server) getImageOwnerHandler(c echo.Context) error {
	// Verify internal auth from worker
	workerSecret := os.Getenv("WORKER_SECRET")
	if workerSecret == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "WORKER_SECRET not configured",
		})
	}
	
	internalAuth := c.Request().Header.Get("X-Internal-Auth")
	if internalAuth != workerSecret {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid internal authentication",
		})
	}

	imageID := c.Param("id")
	if imageID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "image id is required",
		})
	}

	// Validate UUID format
	if _, err := uuid.Parse(imageID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "invalid image id format",
		})
	}

	// Get user ID from header (set by worker after JWT verification)
	requestingUserID := c.Request().Header.Get("X-User-ID")
	if requestingUserID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "X-User-ID header is required",
		})
	}

	// Get image kind (original or staged)
	kind := c.Request().Header.Get("X-Image-Kind")
	if kind == "" {
		kind = "original" // default
	}
	if kind != "original" && kind != "staged" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "X-Image-Kind must be 'original' or 'staged'",
		})
	}

	ctx := c.Request().Context()

	// Fetch image from database
	image, err := s.imageService.GetImageByID(ctx, imageID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "image not found",
		})
	}

	// Get project to determine ownership (images don't have user_id directly)
	// We need to query the projects table via the database
	var projectUserID string
	query := "SELECT user_id FROM projects WHERE id = $1"
	row := s.db.QueryRow(ctx, query, image.ProjectID.String())
	if err := row.Scan(&projectUserID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "failed to determine image ownership",
		})
	}

	// Check ownership
	hasAccess := projectUserID == requestingUserID

	// Build response
	response := ImageOwnershipResponse{
		ImageID:   imageID,
		OwnerID:   projectUserID,
		HasAccess: hasAccess,
	}

	// Only include S3 key if user has access
	if hasAccess {
		// Extract S3 key from the stored URL
		var s3Key string
		bucketName := os.Getenv("S3_BUCKET_NAME")
		if bucketName == "" {
			bucketName = "realstaging-prod"
		}
		
		if kind == "staged" && image.StagedURL != nil && *image.StagedURL != "" {
			s3Key = extractS3KeyFromURL(*image.StagedURL, bucketName)
		} else if kind == "original" && image.OriginalURL != "" {
			s3Key = extractS3KeyFromURL(image.OriginalURL, bucketName)
		}
		response.S3Key = s3Key
	}

	return c.JSON(http.StatusOK, response)
}

// extractS3KeyFromURL extracts the S3 object key from a full S3 URL
// Example: https://s3.us-west-004.backblazeb2.com/bucket-name/path/to/file.jpg -> path/to/file.jpg
func extractS3KeyFromURL(rawURL string, bucketName string) string {
	// For path-style URLs, the pattern is: /bucket-name/key
	prefix := "/" + bucketName + "/"
	if idx := strings.Index(rawURL, prefix); idx >= 0 {
		return rawURL[idx+len(prefix):]
	}
	
	// Fallback: extract everything after last slash
	if idx := strings.LastIndex(rawURL, "/"); idx >= 0 {
		return rawURL[idx+1:]
	}
	
	return ""
}
