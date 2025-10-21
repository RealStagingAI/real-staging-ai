// Package http provides the HTTP server and route handlers.
package http

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	"github.com/real-staging-ai/api/internal/auth"
	"github.com/real-staging-ai/api/internal/billing"
	"github.com/real-staging-ai/api/internal/image"
	"github.com/real-staging-ai/api/internal/logging"
	"github.com/real-staging-ai/api/internal/project"
	"github.com/real-staging-ai/api/internal/reconcile"
	"github.com/real-staging-ai/api/internal/settings"
	"github.com/real-staging-ai/api/internal/sse"
	"github.com/real-staging-ai/api/internal/storage"
	"github.com/real-staging-ai/api/internal/stripe"
	"github.com/real-staging-ai/api/internal/user"
	webdocs "github.com/real-staging-ai/api/web"
)

// Server holds the dependencies for the HTTP server.
type Server struct {
	ctx                 context.Context
	echo                *echo.Echo
	db                  storage.Database
	s3Service           storage.S3Service
	imageService        image.Service
	subscriptionChecker billing.SubscriptionChecker
	authConfig          *auth.Auth0Config
	pubsub              PubSub
}

// NewServer creates and configures a new Echo server.
func NewServer(
	auth0Audience string,
	auth0Domain string,
	ctx context.Context,
	db storage.Database,
	imageService image.Service,
	s3Service storage.S3Service,
	stripeSecretKey string,
) *Server {
	e := echo.New()

	// Add OpenTelemetry middleware
	e.Use(otelecho.Middleware("real-staging-api"))

	// Add other middleware
	e.Use(RequestLoggerMiddleware()) // Custom JSON logger with proper log levels for Render
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods: []string{
			http.MethodGet, http.MethodHead, http.MethodPut,
			http.MethodPatch, http.MethodPost, http.MethodDelete,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization,
		},
	}))

	// Initialize Auth0 config
	authConfig := auth.NewAuth0Config(ctx, auth0Domain, auth0Audience)

	// Initialize billing services
	usageService := billing.NewDefaultUsageService(db)
	subscriptionChecker := billing.NewDefaultSubscriptionChecker(db)

	// Initialize user repository for usage checks
	userRepo := user.NewDefaultRepository(db)

	// Initialize image handler with usage checking
	imgHandler := image.NewDefaultHandler(imageService, usageService, userRepo)

	// Initialize Pub/Sub (Redis) if configured
	var ps PubSub
	if p, err := NewDefaultPubSubFromEnv(); err == nil {
		ps = p
	}

	s := &Server{
		ctx:                 ctx,
		db:                  db,
		s3Service:           s3Service,
		imageService:        imageService,
		subscriptionChecker: subscriptionChecker,
		echo:                e,
		authConfig:          authConfig,
		pubsub:              ps,
	}

	// Health check route
	e.GET("/health", s.healthCheck)

	// Register routes
	api := e.Group("/api/v1")

	// Public routes (no authentication required)
	api.POST("/stripe/webhook", func(c echo.Context) error {
		sh := stripe.NewDefaultHandler(s.db)
		return sh.Webhook(c)
	})

	// Protected routes (require JWT authentication)
	protected := api.Group("")
	protected.Use(auth.JWTMiddleware(s.authConfig))

	// Project routes
	ph := project.NewDefaultHandler(s.db)
	protected.POST("/projects", ph.Create)
	protected.GET("/projects", ph.List)
	protected.GET("/projects/:id", ph.GetByID)
	protected.DELETE("/projects/:id", ph.Delete)

	// Upload routes
	protected.POST("/uploads/presign", s.presignUploadHandler)

	// Image routes
	protected.POST("/images", imgHandler.CreateImage)
	protected.POST("/images/batch", imgHandler.BatchCreateImages)
	protected.GET("/images/:id", imgHandler.GetImage)
	protected.GET("/images/:id/presign", s.presignImageDownloadHandler)
	protected.DELETE("/images/:id", s.deleteImageHandler)
	protected.GET("/projects/:project_id/images", imgHandler.GetProjectImages)
	protected.GET("/projects/:project_id/cost", imgHandler.GetProjectCost)
	
	// Internal routes (for Cloudflare Worker)
	// These use X-Internal-Auth header instead of JWT
	api.GET("/images/:id/owner", s.getImageOwnerHandler)

	// SSE routes
	protected.GET("/events", func(c echo.Context) error {
		cfg := sse.Config{
			SubscribeTimeout: 2000000000,
		}
		h, err := sse.NewDefaultHandlerFromEnv(cfg)
		if err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "pubsub not configured"})
		}
		return h.Events(c)
	})

	// Billing routes
	bh := billing.NewDefaultHandler(s.db, usageService, stripeSecretKey)
	protected.GET("/billing/subscriptions", bh.GetMySubscriptions)
	protected.GET("/billing/invoices", bh.GetMyInvoices)
	protected.GET("/billing/usage", bh.GetMyUsage)
	protected.POST("/billing/create-checkout", bh.CreateCheckoutSession)
	protected.POST("/billing/portal", bh.CreatePortalSession)

	// User profile routes
	profileService := user.NewDefaultProfileService(userRepo)
	profileHandler := NewProfileHandler(profileService, userRepo, logging.Default())
	protected.GET("/user/profile", profileHandler.GetProfile)
	protected.PATCH("/user/profile", profileHandler.UpdateProfile)

	// Admin routes (feature-flagged)
	admin := protected.Group("/admin")
	reconcileSvc := reconcile.NewDefaultService(s.db, s.s3Service)
	reconcileHandler := reconcile.NewDefaultHandler(reconcileSvc)
	admin.POST("/reconcile/images", reconcileHandler.ReconcileImages)
	admin.POST("/reconcile/cleanup-queued", reconcileHandler.CleanupStuckQueuedImages)

	// Admin settings routes
	settingsRepo := settings.NewDefaultRepository(s.db.Pool())
	settingsService := settings.NewDefaultService(settingsRepo)
	adminHandler := NewAdminHandler(settingsService, s.db, logging.Default())
	admin.GET("/models", adminHandler.ListModels)
	admin.GET("/models/active", adminHandler.GetActiveModel)
	admin.PUT("/models/active", adminHandler.UpdateActiveModel)
	admin.GET("/settings", adminHandler.ListSettings)
	admin.GET("/settings/:key", adminHandler.GetSetting)
	admin.PUT("/settings/:key", adminHandler.UpdateSetting)

	// Serve API documentation (embedded)
	webdocs.RegisterRoutes(e)

	return s
}

// NewTestServer creates a new Echo server for testing without Auth0 middleware.
func NewTestServer(
	db storage.Database,
	s3Service storage.S3Service,
	imageService image.Service,
	stripeSecretKey string,
) *Server {
	e := echo.New()

	// Add basic middleware (no Auth0 for testing)
	e.Use(RequestLoggerMiddleware()) // Custom JSON logger with proper log levels for Render
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Initialize billing services for test server
	usageService := billing.NewDefaultUsageService(db)
	userRepo := user.NewDefaultRepository(db)
	subscriptionChecker := billing.NewDefaultSubscriptionChecker(db)

	// Initialize image handler with usage checking
	imgHandler := image.NewDefaultHandler(imageService, usageService, userRepo)

	s := &Server{
		db:                  db,
		s3Service:           s3Service,
		imageService:        imageService,
		subscriptionChecker: subscriptionChecker,
		echo:                e,
		authConfig:          nil,
	}

	// Health check route (same as main server)
	e.GET("/health", s.healthCheck)

	// Register routes without authentication
	api := e.Group("/api/v1")

	// All routes are public for testing
	api.POST("/stripe/webhook", func(c echo.Context) error {
		sh := stripe.NewDefaultHandler(s.db)
		return sh.Webhook(c)
	})

	// Project routes (no auth required for testing)
	ph := project.NewDefaultHandler(s.db)
	api.POST("/projects", withTestUser(ph.Create))
	api.GET("/projects", withTestUser(ph.List))
	api.GET("/projects/:id", withTestUser(ph.GetByID))
	api.PUT("/projects/:id", withTestUser(ph.Update))
	api.DELETE("/projects/:id", withTestUser(ph.Delete))

	// Upload routes
	api.POST("/uploads/presign", withTestUser(s.presignUploadHandler))

	// Image routes
	api.POST("/images", withTestUser(imgHandler.CreateImage))
	api.GET("/images/:id", withTestUser(imgHandler.GetImage))
	api.GET("/images/:id/presign", withTestUser(s.presignImageDownloadHandler))
	api.DELETE("/images/:id", withTestUser(s.deleteImageHandler))
	api.GET("/projects/:project_id/images", withTestUser(imgHandler.GetProjectImages))
	api.GET("/projects/:project_id/cost", withTestUser(imgHandler.GetProjectCost))

	// SSE routes
	api.GET("/events", func(c echo.Context) error {
		cfg := sse.Config{
			SubscribeTimeout: 2000000000,
		}
		h, err := sse.NewDefaultHandlerFromEnv(cfg)
		if err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "pubsub not configured"})
		}
		return h.Events(c)
	})

	// Billing routes (public in test server)
	bh := billing.NewDefaultHandler(s.db, usageService, stripeSecretKey)
	api.GET("/billing/subscriptions", withTestUser(bh.GetMySubscriptions))
	api.GET("/billing/invoices", withTestUser(bh.GetMyInvoices))
	api.GET("/billing/usage", withTestUser(bh.GetMyUsage))
	api.POST("/billing/create-checkout", withTestUser(bh.CreateCheckoutSession))
	api.POST("/billing/portal", withTestUser(bh.CreatePortalSession))

	// User profile routes (test server)
	profileService := user.NewDefaultProfileService(userRepo)
	profileHandler := NewProfileHandler(profileService, userRepo, logging.Default())
	api.GET("/user/profile", withTestUser(profileHandler.GetProfile))
	api.PATCH("/user/profile", withTestUser(profileHandler.UpdateProfile))

	// Admin routes (public in test server, feature-flagged)
	admin := api.Group("/admin")
	reconcileSvc := reconcile.NewDefaultService(s.db, s.s3Service)
	reconcileHandler := reconcile.NewDefaultHandler(reconcileSvc)
	admin.POST("/reconcile/images", reconcileHandler.ReconcileImages)
	admin.POST("/reconcile/cleanup-queued", reconcileHandler.CleanupStuckQueuedImages)

	// Admin settings routes (test server)
	settingsRepo := settings.NewDefaultRepository(s.db.Pool())
	settingsService := settings.NewDefaultService(settingsRepo)
	adminHandler := NewAdminHandler(settingsService, s.db, logging.Default())
	admin.GET("/models", withTestUser(adminHandler.ListModels))
	admin.GET("/models/active", withTestUser(adminHandler.GetActiveModel))
	admin.PUT("/models/active", withTestUser(adminHandler.UpdateActiveModel))
	admin.GET("/settings", withTestUser(adminHandler.ListSettings))
	admin.GET("/settings/:key", withTestUser(adminHandler.GetSetting))
	admin.PUT("/settings/:key", withTestUser(adminHandler.UpdateSetting))

	// Serve API documentation (embedded)
	webdocs.RegisterRoutes(e)

	return s
}

// Start starts the HTTP server.
func (s *Server) Start(addr string) error {
	return s.echo.Start(addr)
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.echo.ServeHTTP(w, r)
}

// healthCheck handles GET /api/v1/health requests.
func (s *Server) healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "real-staging-api",
	})
}

// withTestUser ensures an X-Test-User header is present for test-only servers.
// It defaults to the seeded test user to keep integration tests deterministic.
func withTestUser(h echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Header.Get("X-Test-User") == "" {
			c.Request().Header.Set("X-Test-User", "auth0|testuser")
		}
		return h(c)
	}
}
