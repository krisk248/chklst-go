package handlers

import (
	"chklst-go/internal/database"
	"time"

	"github.com/gofiber/fiber/v3"
)

// HealthCheck returns health status of the application
func HealthCheck(c fiber.Ctx) error {
	// Check database connection
	sqlDB, err := database.DB.DB()
	if err != nil {
		return c.Status(503).JSON(fiber.Map{
			"status":   "unhealthy",
			"database": "disconnected",
			"error":    err.Error(),
		})
	}

	if err := sqlDB.Ping(); err != nil {
		return c.Status(503).JSON(fiber.Map{
			"status":   "unhealthy",
			"database": "unreachable",
			"error":    err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":    "healthy",
		"database":  "connected",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
	})
}
