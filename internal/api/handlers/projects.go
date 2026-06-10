package handlers

import (
	"chklst-go/internal/database"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// uniqueProjectName returns a project name that does not yet exist, appending
// " (copy)", " (copy 2)", ... to the desired base name until a free slot is found.
func uniqueProjectName(tx *gorm.DB, base string) string {
	candidate := base
	for i := 1; ; i++ {
		var count int64
		tx.Model(&database.Project{}).Where("name = ?", candidate).Count(&count)
		if count == 0 {
			return candidate
		}
		if i == 1 {
			candidate = base + " (copy)"
		} else {
			candidate = fmt.Sprintf("%s (copy %d)", base, i)
		}
	}
}

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

	// Reload with components so the response carries the full project. Otherwise
	// the client would replace its state with a component-less project and the
	// component list would appear to vanish until the next refresh.
	database.DB.Preload("Components").First(&project, id)
	return c.JSON(project)
}

// DuplicateProject deep-clones a project (and all its components) under a new name.
// Body: { "new_name": "..." } (optional; defaults to "<name> (copy)").
// The clone is independent — deployments are NOT copied.
func DuplicateProject(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	var source database.Project
	if err := database.DB.Preload("Components").First(&source, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	// Optional new name from body
	var body struct {
		NewName string `json:"new_name"`
	}
	_ = c.Bind().JSON(&body) // body is optional; ignore bind errors

	base := body.NewName
	if base == "" {
		base = source.Name + " (copy)"
	}

	var clone database.Project
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		clone = database.Project{
			Name:           uniqueProjectName(tx, base),
			BuildServer:    source.BuildServer,
			DeployServer:   source.DeployServer,
			DatabaseName:   source.DatabaseName,
			Environment:    source.Environment,
			BackupLocation: source.BackupLocation,
			Description:    source.Description,
		}
		if err := tx.Create(&clone).Error; err != nil {
			return err
		}

		for _, comp := range source.Components {
			newComp := database.Component{
				ProjectID:    clone.ID,
				Name:         comp.Name,
				Developer:    comp.Developer,
				VCSType:      comp.VCSType,
				VCSURL:       comp.VCSURL,
				BuildCommand: comp.BuildCommand,
				ComponentURL: comp.ComponentURL,
				Enabled:      comp.Enabled,
				Description:  comp.Description,
			}
			if err := tx.Create(&newComp).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to duplicate project",
		})
	}

	// Reload with components for the response
	database.DB.Preload("Components").First(&clone, clone.ID)
	return c.Status(201).JSON(clone)
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
