package middleware

import (
	"chklst-go/internal/utils"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// RequestID middleware adds a unique request ID to each request
func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Generate unique request ID
		requestID := uuid.New().String()

		// Store in context
		c.Locals("request_id", requestID)

		// Add to response header
		c.Set("X-Request-ID", requestID)

		// Log request
		utils.AppLogger.WithRequestID(requestID).Info("Incoming request", map[string]interface{}{
			"method": c.Method(),
			"path":   c.Path(),
			"ip":     c.IP(),
		})

		// Continue to next handler
		err := c.Next()

		// Log response
		if err != nil {
			utils.AppLogger.WithRequestID(requestID).Error("Request error", err, map[string]interface{}{
				"method": c.Method(),
				"path":   c.Path(),
				"status": c.Response().StatusCode(),
			})
		}

		return err
	}
}

// GetRequestID retrieves request ID from context
func GetRequestID(c fiber.Ctx) string {
	if requestID, ok := c.Locals("request_id").(string); ok {
		return requestID
	}
	return ""
}
