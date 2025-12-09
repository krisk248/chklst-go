package middleware

import (
	"chklst-go/internal/utils"
	"time"

	"github.com/gofiber/fiber/v3"
)

// RequestLogger logs HTTP requests in structured format
func RequestLogger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get request ID
		requestID := GetRequestID(c)

		// Log request details
		logger := utils.AppLogger
		if requestID != "" {
			reqLogger := logger.WithRequestID(requestID)
			reqLogger.Info("Request completed", map[string]interface{}{
				"method":        c.Method(),
				"path":          c.Path(),
				"status":        c.Response().StatusCode(),
				"duration_ms":   duration.Milliseconds(),
				"ip":            c.IP(),
				"user_agent":    c.Get("User-Agent"),
				"bytes_sent":    len(c.Response().Body()),
			})
		}

		return err
	}
}
