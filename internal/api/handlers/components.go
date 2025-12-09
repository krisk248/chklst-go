package handlers

import (
	"chklst-go/internal/database"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

// CreateComponent creates a new component for a project
func CreateComponent(c fiber.Ctx) error {
	projectID, err := strconv.Atoi(c.Params("projectId"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	var component database.Component
	if err := c.Bind().JSON(&component); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	component.ProjectID = uint(projectID)

	if err := database.DB.Create(&component).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create component",
		})
	}

	return c.Status(201).JSON(component)
}

// UpdateComponent updates an existing component
func UpdateComponent(c fiber.Ctx) error {
	projectID, err := strconv.Atoi(c.Params("projectId"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	componentID, err := strconv.Atoi(c.Params("componentId"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid component ID",
		})
	}

	var component database.Component
	if err := database.DB.First(&component, componentID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Component not found",
		})
	}

	if component.ProjectID != uint(projectID) {
		return c.Status(400).JSON(fiber.Map{
			"error": "Component does not belong to this project",
		})
	}

	if err := c.Bind().JSON(&component); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := database.DB.Save(&component).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update component",
		})
	}

	return c.JSON(component)
}

// DeleteComponent deletes a component
func DeleteComponent(c fiber.Ctx) error {
	componentID, err := strconv.Atoi(c.Params("componentId"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid component ID",
		})
	}

	if err := database.DB.Delete(&database.Component{}, componentID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete component",
		})
	}

	return c.SendStatus(204)
}
