package handlers

import (
	"chklst-go/internal/database"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

// ListProjects returns all projects
func ListProjects(c fiber.Ctx) error {
	var projects []database.Project
	
	// Preload components
	if err := database.DB.Preload("Components").Find(&projects).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch projects",
		})
	}

	return c.JSON(projects)
}

// GetProject returns a single project by ID
func GetProject(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	var project database.Project
	if err := database.DB.Preload("Components").First(&project, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	return c.JSON(project)
}

// CreateProject creates a new project
func CreateProject(c fiber.Ctx) error {
	var project database.Project

	if err := c.Bind().JSON(&project); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if project.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Project name is required",
		})
	}

	if err := database.DB.Create(&project).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create project",
		})
	}

	return c.Status(201).JSON(project)
}

// UpdateProject updates an existing project
func UpdateProject(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	var project database.Project
	if err := database.DB.First(&project, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	if err := c.Bind().JSON(&project); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := database.DB.Save(&project).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update project",
		})
	}

	return c.JSON(project)
}

// DeleteProject deletes a project
func DeleteProject(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	if err := database.DB.Delete(&database.Project{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete project",
		})
	}

	return c.SendStatus(204)
}
