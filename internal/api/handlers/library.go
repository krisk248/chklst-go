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

// UpdateLibrary updates the entire library (bulk update)
func UpdateLibrary(c fiber.Ctx) error {
	var req database.Library
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

	// Update all fields
	library.Developers = req.Developers
	library.BuildServers = req.BuildServers
	library.DeployServers = req.DeployServers
	library.Environments = req.Environments

	if err := database.DB.Save(&library).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update library",
		})
	}

	return c.JSON(library)
}

// AddBuildServer adds a build server to the library
func AddBuildServer(c fiber.Ctx) error {
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

	// Check for duplicate
	for _, server := range library.BuildServers {
		if server == req.Name {
			return c.Status(409).JSON(fiber.Map{
				"error": "Build server already exists",
			})
		}
	}

	library.BuildServers = append(library.BuildServers, req.Name)
	if err := database.DB.Save(&library).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update library",
		})
	}

	return c.Status(201).JSON(library)
}

// RemoveBuildServer removes a build server from the library
func RemoveBuildServer(c fiber.Ctx) error {
	name := c.Params("name")

	var library database.Library
	if err := database.DB.First(&library).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch library",
		})
	}

	newServers := []string{}
	for _, server := range library.BuildServers {
		if server != name {
			newServers = append(newServers, server)
		}
	}

	library.BuildServers = newServers
	if err := database.DB.Save(&library).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update library",
		})
	}

	return c.SendStatus(204)
}

// AddDeployServer adds a deploy server to the library
func AddDeployServer(c fiber.Ctx) error {
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

	for _, server := range library.DeployServers {
		if server == req.Name {
			return c.Status(409).JSON(fiber.Map{
				"error": "Deploy server already exists",
			})
		}
	}

	library.DeployServers = append(library.DeployServers, req.Name)
	if err := database.DB.Save(&library).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update library",
		})
	}

	return c.Status(201).JSON(library)
}

// RemoveDeployServer removes a deploy server from the library
func RemoveDeployServer(c fiber.Ctx) error {
	name := c.Params("name")

	var library database.Library
	if err := database.DB.First(&library).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch library",
		})
	}

	newServers := []string{}
	for _, server := range library.DeployServers {
		if server != name {
			newServers = append(newServers, server)
		}
	}

	library.DeployServers = newServers
	if err := database.DB.Save(&library).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update library",
		})
	}

	return c.SendStatus(204)
}

// AddEnvironment adds an environment to the library
func AddEnvironment(c fiber.Ctx) error {
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

	for _, env := range library.Environments {
		if env == req.Name {
			return c.Status(409).JSON(fiber.Map{
				"error": "Environment already exists",
			})
		}
	}

	library.Environments = append(library.Environments, req.Name)
	if err := database.DB.Save(&library).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update library",
		})
	}

	return c.Status(201).JSON(library)
}

// RemoveEnvironment removes an environment from the library
func RemoveEnvironment(c fiber.Ctx) error {
	name := c.Params("name")

	var library database.Library
	if err := database.DB.First(&library).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch library",
		})
	}

	newEnvs := []string{}
	for _, env := range library.Environments {
		if env != name {
			newEnvs = append(newEnvs, env)
		}
	}

	library.Environments = newEnvs
	if err := database.DB.Save(&library).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update library",
		})
	}

	return c.SendStatus(204)
}
