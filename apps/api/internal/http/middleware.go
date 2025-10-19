package http

import (
	"time"

	"github.com/labstack/echo/v4"

	"github.com/real-staging-ai/api/internal/logging"
)

// RequestLoggerMiddleware logs HTTP requests in JSON format with proper log levels.
// This ensures Render correctly interprets log levels instead of marking everything as errors.
func RequestLoggerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			res := c.Response()

			// Process request
			err := next(c)
			if err != nil {
				c.Error(err)
			}

			// Calculate latency
			latency := time.Since(start)
			status := res.Status

			// Determine log level based on status code
			log := logging.Default()
			ctx := req.Context()

			// Build log fields
			fields := []any{
				"method", req.Method,
				"uri", req.RequestURI,
				"status", status,
				"latency_ms", latency.Milliseconds(),
				"remote_ip", c.RealIP(),
				"user_agent", req.UserAgent(),
				"bytes_in", req.Header.Get(echo.HeaderContentLength),
				"bytes_out", res.Size,
			}

			// Add error if present
			if err != nil {
				fields = append(fields, "error", err.Error())
			}

			// Log at appropriate level
			switch {
			case status >= 500:
				log.Error(ctx, "HTTP request", fields...)
			case status >= 400:
				log.Warn(ctx, "HTTP request", fields...)
			default:
				log.Info(ctx, "HTTP request", fields...)
			}

			return nil
		}
	}
}
