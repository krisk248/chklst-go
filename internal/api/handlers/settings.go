package handlers

import (
	"chklst-go/internal/database"

	"github.com/gofiber/fiber/v3"
)

// GetSettings returns the application settings
func GetSettings(c fiber.Ctx) error {
	var settings database.Settings

	// Get or create settings (singleton pattern)
	if err := database.DB.First(&settings).Error; err != nil {
		// Create default settings if not exists
		settings = database.Settings{
			DefaultDeployedBy:  "Kannan",
			ExcelExportPath:    "/reports",
			AutoClearAfterSave: false,
		}
		if err := database.DB.Create(&settings).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to create default settings",
			})
		}
	}

	return c.JSON(settings)
}

// UpdateSettings updates the application settings
func UpdateSettings(c fiber.Ctx) error {
	var req database.Settings
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var settings database.Settings
	if err := database.DB.First(&settings).Error; err != nil {
		// Create if not exists
		settings = req
		settings.ID = 1
		if err := database.DB.Create(&settings).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to create settings",
			})
		}
		return c.JSON(settings)
	}

	// Update existing settings
	settings.DefaultDeployedBy = req.DefaultDeployedBy
	settings.ExcelExportPath = req.ExcelExportPath
	settings.AutoClearAfterSave = req.AutoClearAfterSave

	if err := database.DB.Save(&settings).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update settings",
		})
	}

	return c.JSON(settings)
}
