package handlers

import (
	"chklst-go/internal/database"

	"github.com/gofiber/fiber/v3"
)

// GetLibrary returns the library presets
func GetLibrary(c fiber.Ctx) error {
	var library database.Library
	
	if err := database.DB.First(&library).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch library",
		})
	}

	return c.JSON(library)
}

// AddDeveloper adds a developer to the library
func AddDeveloper(c fiber.Ctx) error {
	var req struct {
		Name string `json:"name"`
	}

	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var library database.Library
	if err := database.DB.First(&library).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch library",
		})
	}

	// Add developer if not exists
	for _, dev := range library.Developers {
		if dev == req.Name {
			return c.Status(409).JSON(fiber.Map{
				"error": "Developer already exists",
			})
		}
	}

	library.Developers = append(library.Developers, req.Name)
	if err := database.DB.Save(&library).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update library",
		})
	}

	return c.Status(201).JSON(library)
}

// RemoveDeveloper removes a developer from the library
func RemoveDeveloper(c fiber.Ctx) error {
	name := c.Params("name")

	var library database.Library
	if err := database.DB.First(&library).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch library",
		})
	}

	// Remove developer
	newDevs := []string{}
	for _, dev := range library.Developers {
		if dev != name {
			newDevs = append(newDevs, dev)
		}
	}

	library.Developers = newDevs
	if err := database.DB.Save(&library).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update library",
		})
	}

	return c.SendStatus(204)
}

// Similar functions for BuildServers, DeployServers, Environments...
// (Omitted for brevity but follow same pattern)
