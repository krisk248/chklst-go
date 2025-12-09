package middleware

import (
	"chklst-go/internal/utils"
	"runtime/debug"

	"github.com/gofiber/fiber/v3"
)

// Recovery middleware recovers from panics
func Recovery() fiber.Handler {
	return func(c fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				requestID := GetRequestID(c)
				stackTrace := string(debug.Stack())

				utils.AppLogger.WithRequestID(requestID).Error("Panic recovered", nil, map[string]interface{}{
					"panic":       r,
					"stack_trace": stackTrace,
					"method":      c.Method(),
					"path":        c.Path(),
				})

				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":      "Internal server error",
					"request_id": requestID,
				})
			}
		}()

		return c.Next()
	}
}
