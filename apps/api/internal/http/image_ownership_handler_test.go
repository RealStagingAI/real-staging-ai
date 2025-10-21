package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/real-staging-ai/api/internal/image"
	"github.com/real-staging-ai/api/internal/storage"
)

func TestGetImageOwnerHandler(t *testing.T) {
	// Set worker secret for tests
	t.Setenv("WORKER_SECRET", "test-worker-secret")

	imageID := uuid.New()
	projectID := uuid.New()
	userID := uuid.New()
	otherUserID := uuid.New()

	t.Run("success: returns ownership info with s3 key for owned original image", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/v1/images/"+imageID.String()+"/owner", nil)
		req.Header.Set("X-Internal-Auth", "test-worker-secret")
		req.Header.Set("X-User-ID", userID.String())
		req.Header.Set("X-Image-Kind", "original")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/v1/images/:id/owner")
		c.SetParamNames("id")
		c.SetParamValues(imageID.String())

		// Mock image service
		mockImageService := &image.ServiceMock{
			GetImageByIDFunc: func(ctx context.Context, id string) (*image.Image, error) {
				assert.Equal(t, imageID.String(), id)
				return &image.Image{
					ID:          imageID,
					ProjectID:   projectID,
					OriginalURL: "https://s3.us-west-004.backblazeb2.com/realstaging-prod/images/test/original.jpg",
					Status:      image.StatusReady,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}, nil
			},
		}

		// Mock database
		mockDB := &storage.DatabaseMock{
			QueryRowFunc: func(ctx context.Context, query string, args ...interface{}) pgx.Row {
				assert.Contains(t, query, "SELECT user_id FROM projects")
				assert.Equal(t, projectID.String(), args[0])
				return &mockRow{
					scanFunc: func(dest ...interface{}) error {
						if userIDPtr, ok := dest[0].(*string); ok {
							*userIDPtr = userID.String()
						}
						return nil
					},
				}
			},
		}

		server := &Server{
			imageService: mockImageService,
			db:           mockDB,
		}

		err := server.getImageOwnerHandler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Verify response
		assert.Contains(t, rec.Body.String(), `"image_id":"`)
		assert.Contains(t, rec.Body.String(), `"owner_id":"`)
		assert.Contains(t, rec.Body.String(), `"has_access":true`)
		assert.Contains(t, rec.Body.String(), `"s3_key":"images/test/original.jpg"`)
	})

	t.Run("success: returns ownership info with s3 key for owned staged image", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/v1/images/"+imageID.String()+"/owner", nil)
		req.Header.Set("X-Internal-Auth", "test-worker-secret")
		req.Header.Set("X-User-ID", userID.String())
		req.Header.Set("X-Image-Kind", "staged")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/v1/images/:id/owner")
		c.SetParamNames("id")
		c.SetParamValues(imageID.String())

		stagedURL := "https://s3.us-west-004.backblazeb2.com/realstaging-prod/images/test/staged.jpg"

		mockImageService := &image.ServiceMock{
			GetImageByIDFunc: func(ctx context.Context, id string) (*image.Image, error) {
				return &image.Image{
					ID:          imageID,
					ProjectID:   projectID,
					OriginalURL: "https://s3.us-west-004.backblazeb2.com/realstaging-prod/images/test/original.jpg",
					StagedURL:   &stagedURL,
					Status:      image.StatusReady,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}, nil
			},
		}

		mockDB := &storage.DatabaseMock{
			QueryRowFunc: func(ctx context.Context, query string, args ...interface{}) pgx.Row {
				return &mockRow{
					scanFunc: func(dest ...interface{}) error {
						if userIDPtr, ok := dest[0].(*string); ok {
							*userIDPtr = userID.String()
						}
						return nil
					},
				}
			},
		}

		server := &Server{
			imageService: mockImageService,
			db:           mockDB,
		}

		err := server.getImageOwnerHandler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"has_access":true`)
		assert.Contains(t, rec.Body.String(), `"s3_key":"images/test/staged.jpg"`)
	})

	t.Run("success: returns has_access false for non-owned image", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/v1/images/"+imageID.String()+"/owner", nil)
		req.Header.Set("X-Internal-Auth", "test-worker-secret")
		req.Header.Set("X-User-ID", otherUserID.String())
		req.Header.Set("X-Image-Kind", "original")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/v1/images/:id/owner")
		c.SetParamNames("id")
		c.SetParamValues(imageID.String())

		mockImageService := &image.ServiceMock{
			GetImageByIDFunc: func(ctx context.Context, id string) (*image.Image, error) {
				return &image.Image{
					ID:          imageID,
					ProjectID:   projectID,
					OriginalURL: "https://s3.us-west-004.backblazeb2.com/realstaging-prod/images/test/original.jpg",
					Status:      image.StatusReady,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}, nil
			},
		}

		mockDB := &storage.DatabaseMock{
			QueryRowFunc: func(ctx context.Context, query string, args ...interface{}) pgx.Row {
				return &mockRow{
					scanFunc: func(dest ...interface{}) error {
						if userIDPtr, ok := dest[0].(*string); ok {
							*userIDPtr = userID.String() // Different from requesting user
						}
						return nil
					},
				}
			},
		}

		server := &Server{
			imageService: mockImageService,
			db:           mockDB,
		}

		err := server.getImageOwnerHandler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"has_access":false`)
		assert.NotContains(t, rec.Body.String(), `"s3_key"`) // Should not include S3 key for non-owned images
	})

	t.Run("fail: missing worker secret", func(t *testing.T) {
		t.Setenv("WORKER_SECRET", "")

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/v1/images/"+imageID.String()+"/owner", nil)
		req.Header.Set("X-Internal-Auth", "test-worker-secret")
		req.Header.Set("X-User-ID", userID.String())
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/v1/images/:id/owner")
		c.SetParamNames("id")
		c.SetParamValues(imageID.String())

		server := &Server{}

		err := server.getImageOwnerHandler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "WORKER_SECRET not configured")
	})

	t.Run("fail: invalid internal auth", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/v1/images/"+imageID.String()+"/owner", nil)
		req.Header.Set("X-Internal-Auth", "wrong-secret")
		req.Header.Set("X-User-ID", userID.String())
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/v1/images/:id/owner")
		c.SetParamNames("id")
		c.SetParamValues(imageID.String())

		server := &Server{}

		err := server.getImageOwnerHandler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid internal authentication")
	})

	t.Run("fail: missing image id", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/v1/images//owner", nil)
		req.Header.Set("X-Internal-Auth", "test-worker-secret")
		req.Header.Set("X-User-ID", userID.String())
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/v1/images/:id/owner")
		c.SetParamNames("id")
		c.SetParamValues("")

		server := &Server{}

		err := server.getImageOwnerHandler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "image id is required")
	})

	t.Run("fail: invalid image id format", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/v1/images/invalid-uuid/owner", nil)
		req.Header.Set("X-Internal-Auth", "test-worker-secret")
		req.Header.Set("X-User-ID", userID.String())
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/v1/images/:id/owner")
		c.SetParamNames("id")
		c.SetParamValues("invalid-uuid")

		server := &Server{}

		err := server.getImageOwnerHandler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid image id format")
	})

	t.Run("fail: missing user id header", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/v1/images/"+imageID.String()+"/owner", nil)
		req.Header.Set("X-Internal-Auth", "test-worker-secret")
		// Missing X-User-ID header
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/v1/images/:id/owner")
		c.SetParamNames("id")
		c.SetParamValues(imageID.String())

		server := &Server{}

		err := server.getImageOwnerHandler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "X-User-ID header is required")
	})

	t.Run("fail: invalid image kind", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/v1/images/"+imageID.String()+"/owner", nil)
		req.Header.Set("X-Internal-Auth", "test-worker-secret")
		req.Header.Set("X-User-ID", userID.String())
		req.Header.Set("X-Image-Kind", "invalid")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/v1/images/:id/owner")
		c.SetParamNames("id")
		c.SetParamValues(imageID.String())

		server := &Server{}

		err := server.getImageOwnerHandler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "X-Image-Kind must be 'original' or 'staged'")
	})

	t.Run("fail: image not found", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/v1/images/"+imageID.String()+"/owner", nil)
		req.Header.Set("X-Internal-Auth", "test-worker-secret")
		req.Header.Set("X-User-ID", userID.String())
		req.Header.Set("X-Image-Kind", "original")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/v1/images/:id/owner")
		c.SetParamNames("id")
		c.SetParamValues(imageID.String())

		mockImageService := &image.ServiceMock{
			GetImageByIDFunc: func(ctx context.Context, id string) (*image.Image, error) {
				return nil, errors.New("not found")
			},
		}

		server := &Server{
			imageService: mockImageService,
		}

		err := server.getImageOwnerHandler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "image not found")
	})

	t.Run("fail: project ownership query fails", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/v1/images/"+imageID.String()+"/owner", nil)
		req.Header.Set("X-Internal-Auth", "test-worker-secret")
		req.Header.Set("X-User-ID", userID.String())
		req.Header.Set("X-Image-Kind", "original")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/v1/images/:id/owner")
		c.SetParamNames("id")
		c.SetParamValues(imageID.String())

		mockImageService := &image.ServiceMock{
			GetImageByIDFunc: func(ctx context.Context, id string) (*image.Image, error) {
				return &image.Image{
					ID:          imageID,
					ProjectID:   projectID,
					OriginalURL: "https://s3.us-west-004.backblazeb2.com/realstaging-prod/images/test/original.jpg",
					Status:      image.StatusReady,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}, nil
			},
		}

		mockDB := &storage.DatabaseMock{
			QueryRowFunc: func(ctx context.Context, query string, args ...interface{}) pgx.Row {
				return &mockRow{
					scanFunc: func(dest ...interface{}) error {
						return errors.New("database error")
					},
				}
			},
		}

		server := &Server{
			imageService: mockImageService,
			db:           mockDB,
		}

		err := server.getImageOwnerHandler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "failed to determine image ownership")
	})
}

func TestExtractS3KeyFromURL(t *testing.T) {
	tests := []struct {
		name       string
		rawURL     string
		bucketName string
		expected   string
	}{
		{
			name:       "success: path-style URL",
			rawURL:     "https://s3.us-west-004.backblazeb2.com/realstaging-prod/images/abc123/original.jpg",
			bucketName: "realstaging-prod",
			expected:   "images/abc123/original.jpg",
		},
		{
			name:       "success: path-style URL with subdirectories",
			rawURL:     "https://s3.us-west-004.backblazeb2.com/realstaging-prod/path/to/image/file.jpg",
			bucketName: "realstaging-prod",
			expected:   "path/to/image/file.jpg",
		},
		{
			name:       "success: fallback to last slash when bucket not found",
			rawURL:     "https://cdn.example.com/some/path/file.jpg",
			bucketName: "unknown-bucket",
			expected:   "file.jpg",
		},
		{
			name:       "success: handles different bucket names",
			rawURL:     "https://s3.us-west-004.backblazeb2.com/my-bucket/images/test.jpg",
			bucketName: "my-bucket",
			expected:   "images/test.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractS3KeyFromURL(tt.rawURL, tt.bucketName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// mockRow implements pgx.Row interface for testing
type mockRow struct {
	scanFunc func(dest ...interface{}) error
}

func (m *mockRow) Scan(dest ...interface{}) error {
	return m.scanFunc(dest...)
}
