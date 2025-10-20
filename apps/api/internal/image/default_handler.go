package image

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/real-staging-ai/api/internal/auth"
	"github.com/real-staging-ai/api/internal/user"
)

//go:generate go run github.com/matryer/moq@v0.5.3 -out usage_checker_mock.go . UsageChecker

// UsageChecker provides methods to check if a user can create images.
type UsageChecker interface {
	CanCreateImage(ctx context.Context, userID string) (bool, error)
}

// DefaultHandler contains the HTTP handlers for image operations.
type DefaultHandler struct {
	service      Service
	usageChecker UsageChecker
	userRepo     user.Repository
}

// NewDefaultHandler creates a new Handler instance.
func NewDefaultHandler(service Service, usageChecker UsageChecker, userRepo user.Repository) *DefaultHandler {
	return &DefaultHandler{
		service:      service,
		usageChecker: usageChecker,
		userRepo:     userRepo,
	}
}

// CreateImage handles POST /api/v1/images requests.
func (h *DefaultHandler) CreateImage(c echo.Context) error {
	var req CreateImageRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request format",
		})
	}

	// Validate request
	if validationErrs := h.validateCreateImageRequest(&req); len(validationErrs) > 0 {
		return c.JSON(http.StatusUnprocessableEntity, ValidationErrorResponse{
			Error:            "validation_failed",
			Message:          "The provided data is invalid",
			ValidationErrors: validationErrs,
		})
	}

	// Check usage limits if usage checker is configured
	if h.usageChecker != nil && h.userRepo != nil {
		// Get user ID from auth
		auth0Sub, err := auth.GetUserIDOrDefault(c)
		if err == nil && auth0Sub != "" {
			// Get user from database
			userRow, err := h.userRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
			if err == nil {
				// Check if user can create image
				canCreate, err := h.usageChecker.CanCreateImage(c.Request().Context(), userRow.ID.String())
				if err == nil && !canCreate {
					return c.JSON(http.StatusPaymentRequired, ErrorResponse{
						Error:   "usage_limit_exceeded",
						Message: "You have reached your monthly image limit. Please upgrade your plan to continue.",
					})
				}
			}
		}
	}

	// Create the image
	img, err := h.service.CreateImage(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to create image",
		})
	}

	return c.JSON(http.StatusCreated, img)
}

// BatchCreateImages handles POST /api/v1/images/batch requests.
func (h *DefaultHandler) BatchCreateImages(c echo.Context) error {
	var req BatchCreateImagesRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request format",
		})
	}

	// Validate batch request
	if len(req.Images) == 0 {
		return c.JSON(http.StatusUnprocessableEntity, ValidationErrorResponse{
			Error:   "validation_failed",
			Message: "At least one image is required",
			ValidationErrors: []ValidationErrorDetail{{
				Field:   "images",
				Message: "images array cannot be empty",
			}},
		})
	}

	if len(req.Images) > 50 {
		return c.JSON(http.StatusUnprocessableEntity, ValidationErrorResponse{
			Error:   "validation_failed",
			Message: "Too many images",
			ValidationErrors: []ValidationErrorDetail{{
				Field:   "images",
				Message: "maximum 50 images per batch request",
			}},
		})
	}

	// Validate each image request
	var allValidationErrors []ValidationErrorDetail
	for i, imgReq := range req.Images {
		if errors := h.validateCreateImageRequest(&imgReq); len(errors) > 0 {
			for _, err := range errors {
				allValidationErrors = append(allValidationErrors, ValidationErrorDetail{
					Field:   fmt.Sprintf("images[%d].%s", i, err.Field),
					Message: err.Message,
				})
			}
		}
	}

	if len(allValidationErrors) > 0 {
		return c.JSON(http.StatusUnprocessableEntity, ValidationErrorResponse{
			Error:            "validation_failed",
			Message:          "One or more images have invalid data",
			ValidationErrors: allValidationErrors,
		})
	}

	// Check usage limits for batch if usage checker is configured
	if h.usageChecker != nil && h.userRepo != nil {
		// Get user ID from auth
		auth0Sub, err := auth.GetUserIDOrDefault(c)
		if err == nil && auth0Sub != "" {
			// Get user from database
			userRow, err := h.userRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
			if err == nil {
				// Check if user can create image (checks overall limit)
				canCreate, err := h.usageChecker.CanCreateImage(c.Request().Context(), userRow.ID.String())
				if err == nil && !canCreate {
					return c.JSON(http.StatusPaymentRequired, ErrorResponse{
						Error:   "usage_limit_exceeded",
						Message: "You have reached your monthly image limit. Please upgrade your plan to continue.",
					})
				}
			}
		}
	}

	// Create the images in batch
	response, err := h.service.BatchCreateImages(c.Request().Context(), req.Images)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to create images",
		})
	}

	// Return 207 Multi-Status if partial success, 201 if all success
	statusCode := http.StatusCreated
	if response.Failed > 0 && response.Success > 0 {
		statusCode = http.StatusMultiStatus
	} else if response.Failed > 0 {
		statusCode = http.StatusBadRequest
	}

	return c.JSON(statusCode, response)
}

// GetImage handles GET /api/v1/images/{id} requests.
func (h *DefaultHandler) GetImage(c echo.Context) error {
	imageID := c.Param("id")
	if imageID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Image ID is required",
		})
	}

	// Validate UUID format
	if _, err := uuid.Parse(imageID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid image ID format",
		})
	}

	img, err := h.service.GetImageByID(c.Request().Context(), imageID)
	if err != nil {
		// Check if it's a not found error
		if err.Error() == "no rows in result set" {
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "Image not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to get image",
		})
	}

	return c.JSON(http.StatusOK, img)
}

// GetProjectImages handles GET /api/v1/projects/{project_id}/images requests.
func (h *DefaultHandler) GetProjectImages(c echo.Context) error {
	projectID := c.Param("project_id")
	if projectID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Project ID is required",
		})
	}

	// Validate UUID format
	if _, err := uuid.Parse(projectID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid project ID format",
		})
	}

	images, err := h.service.GetImagesByProjectID(c.Request().Context(), projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to get images",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"images": images,
	})
}

// DeleteImage handles DELETE /api/v1/images/{id} requests.
func (h *DefaultHandler) DeleteImage(c echo.Context) error {
	imageID := c.Param("id")
	if imageID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Image ID is required",
		})
	}

	// Validate UUID format
	if _, err := uuid.Parse(imageID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid image ID format",
		})
	}

	err := h.service.DeleteImage(c.Request().Context(), imageID)
	if err != nil {
		// Check if it's a not found error
		if err.Error() == "no rows in result set" {
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "Image not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to delete image",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// validateCreateImageRequest validates the create image request.
func (h *DefaultHandler) validateCreateImageRequest(req *CreateImageRequest) []ValidationErrorDetail {
	var errors []ValidationErrorDetail

	// Validate project ID
	if req.ProjectID == uuid.Nil {
		errors = append(errors, ValidationErrorDetail{
			Field:   "project_id",
			Message: "project_id is required",
		})
	}

	// Validate original URL
	if req.OriginalURL == "" {
		errors = append(errors, ValidationErrorDetail{
			Field:   "original_url",
			Message: "original_url is required",
		})
	}

	// Validate room type if provided
	if req.RoomType != nil {
		validRoomTypes := []string{"living_room", "bedroom", "kitchen", "bathroom",
			"dining_room", "office", "entryway", "outdoor"}
		isValid := slices.Contains(validRoomTypes, *req.RoomType)
		if !isValid {
			errors = append(errors, ValidationErrorDetail{
				Field: "room_type",
				Message: "room_type must be one of: " +
					"living_room, bedroom, kitchen, bathroom, dining_room, office, entryway, outdoor",
			})
		}
	}

	// Validate style if provided
	if req.Style != nil {
		validStyles := []string{"modern", "contemporary", "traditional", "industrial", "scandinavian"}
		isValid := slices.Contains(validStyles, *req.Style)
		if !isValid {
			errors = append(errors, ValidationErrorDetail{
				Field:   "style",
				Message: "style must be one of: modern, contemporary, traditional, industrial, scandinavian",
			})
		}
	}

	// Validate seed if provided
	if req.Seed != nil {
		if *req.Seed < 1 || *req.Seed > 4294967295 {
			errors = append(errors, ValidationErrorDetail{
				Field:   "seed",
				Message: "seed must be between 1 and 4294967295",
			})
		}
	}

	return errors
}

// GetProjectCost handles GET /api/v1/projects/:project_id/cost requests.
func (h *DefaultHandler) GetProjectCost(c echo.Context) error {
	projectID := c.Param("project_id")

	// Validate project ID format
	if _, err := uuid.Parse(projectID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid project ID format",
		})
	}

	// Get cost summary
	summary, err := h.service.GetProjectCostSummary(c.Request().Context(), projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to retrieve cost summary",
		})
	}

	return c.JSON(http.StatusOK, summary)
}
