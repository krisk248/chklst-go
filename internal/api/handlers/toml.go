package handlers

import (
	"chklst-go/internal/database"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/pelletier/go-toml/v2"
	"gorm.io/gorm"
)

// ========== LIBRARY TOML ==========

// LibraryTOML represents library data in TOML format
type LibraryTOML struct {
	Library LibraryData `toml:"library"`
}

type LibraryData struct {
	Developers    []string `toml:"developers"`
	BuildServers  []string `toml:"build_servers"`
	DeployServers []string `toml:"deploy_servers"`
	Environments  []string `toml:"environments"`
}

// ExportLibraryTOML exports library as TOML
func ExportLibraryTOML(c fiber.Ctx) error {
	var library database.Library
	if err := database.DB.First(&library).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch library",
		})
	}

	data := LibraryTOML{
		Library: LibraryData{
			Developers:    library.Developers,
			BuildServers:  library.BuildServers,
			DeployServers: library.DeployServers,
			Environments:  library.Environments,
		},
	}

	tomlBytes, err := toml.Marshal(data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to generate TOML",
		})
	}

	c.Set("Content-Type", "application/toml")
	c.Set("Content-Disposition", "attachment; filename=library.toml")
	return c.Send(tomlBytes)
}

// ImportLibraryRequest represents import request
type ImportLibraryRequest struct {
	Content string `json:"content"` // TOML content as string
	Mode    string `json:"mode"`    // "merge", "skip", "overwrite"
}

// ImportLibraryTOML imports library from TOML with conflict handling
func ImportLibraryTOML(c fiber.Ctx) error {
	var req ImportLibraryRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Parse TOML
	var data LibraryTOML
	if err := toml.Unmarshal([]byte(req.Content), &data); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid TOML format: " + err.Error(),
		})
	}

	// Get existing library
	var library database.Library
	if err := database.DB.First(&library).Error; err != nil {
		// Create new if not exists
		library = database.Library{ID: 1}
	}

	mode := req.Mode
	if mode == "" {
		mode = "merge"
	}

	switch mode {
	case "overwrite":
		library.Developers = data.Library.Developers
		library.BuildServers = data.Library.BuildServers
		library.DeployServers = data.Library.DeployServers
		library.Environments = data.Library.Environments
	case "merge":
		library.Developers = mergeStringSlices(library.Developers, data.Library.Developers)
		library.BuildServers = mergeStringSlices(library.BuildServers, data.Library.BuildServers)
		library.DeployServers = mergeStringSlices(library.DeployServers, data.Library.DeployServers)
		library.Environments = mergeStringSlices(library.Environments, data.Library.Environments)
	case "skip":
		// Only add items that don't exist
		library.Developers = mergeStringSlices(library.Developers, data.Library.Developers)
		library.BuildServers = mergeStringSlices(library.BuildServers, data.Library.BuildServers)
		library.DeployServers = mergeStringSlices(library.DeployServers, data.Library.DeployServers)
		library.Environments = mergeStringSlices(library.Environments, data.Library.Environments)
	}

	if err := database.DB.Save(&library).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save library",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Library imported successfully",
		"library": library,
	})
}

// ========== PROJECTS TOML ==========

// ProjectsTOML represents projects data in TOML format
type ProjectsTOML struct {
	Projects []ProjectData `toml:"projects"`
}

type ProjectData struct {
	Name           string          `toml:"name"`
	BuildServer    string          `toml:"build_server,omitempty"`
	DeployServer   string          `toml:"deploy_server,omitempty"`
	DatabaseName   string          `toml:"database_name,omitempty"`
	Environment    string          `toml:"environment,omitempty"`
	BackupLocation string          `toml:"backup_location,omitempty"`
	Description    string          `toml:"description,omitempty"`
	Components     []ComponentData `toml:"components,omitempty"`
}

type ComponentData struct {
	Name         string `toml:"name"`
	Developer    string `toml:"developer,omitempty"`
	VCSType      string `toml:"vcs_type,omitempty"`
	VCSURL       string `toml:"vcs_url,omitempty"`
	BuildCommand string `toml:"build_command,omitempty"`
	ComponentURL string `toml:"component_url,omitempty"`
	Enabled      bool   `toml:"enabled"`
	Description  string `toml:"description,omitempty"`
}

// ExportProjectsTOML exports all projects as TOML
func ExportProjectsTOML(c fiber.Ctx) error {
	var projects []database.Project
	if err := database.DB.Preload("Components").Find(&projects).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch projects",
		})
	}

	data := ProjectsTOML{
		Projects: make([]ProjectData, len(projects)),
	}

	for i, p := range projects {
		data.Projects[i] = ProjectData{
			Name:           p.Name,
			BuildServer:    p.BuildServer,
			DeployServer:   p.DeployServer,
			DatabaseName:   p.DatabaseName,
			Environment:    p.Environment,
			BackupLocation: p.BackupLocation,
			Description:    p.Description,
			Components:     make([]ComponentData, len(p.Components)),
		}

		for j, comp := range p.Components {
			data.Projects[i].Components[j] = ComponentData{
				Name:         comp.Name,
				Developer:    comp.Developer,
				VCSType:      comp.VCSType,
				VCSURL:       comp.VCSURL,
				BuildCommand: comp.BuildCommand,
				ComponentURL: comp.ComponentURL,
				Enabled:      comp.Enabled,
				Description:  comp.Description,
			}
		}
	}

	tomlBytes, err := toml.Marshal(data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to generate TOML",
		})
	}

	c.Set("Content-Type", "application/toml")
	c.Set("Content-Disposition", "attachment; filename=projects.toml")
	return c.Send(tomlBytes)
}

// ExportSingleProjectTOML exports a single project as TOML
func ExportSingleProjectTOML(c fiber.Ctx) error {
	id := c.Params("id")

	var project database.Project
	if err := database.DB.Preload("Components").First(&project, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	data := ProjectsTOML{
		Projects: []ProjectData{{
			Name:           project.Name,
			BuildServer:    project.BuildServer,
			DeployServer:   project.DeployServer,
			DatabaseName:   project.DatabaseName,
			Environment:    project.Environment,
			BackupLocation: project.BackupLocation,
			Description:    project.Description,
			Components:     make([]ComponentData, len(project.Components)),
		}},
	}

	for j, comp := range project.Components {
		data.Projects[0].Components[j] = ComponentData{
			Name:         comp.Name,
			Developer:    comp.Developer,
			VCSType:      comp.VCSType,
			VCSURL:       comp.VCSURL,
			BuildCommand: comp.BuildCommand,
			ComponentURL: comp.ComponentURL,
			Enabled:      comp.Enabled,
			Description:  comp.Description,
		}
	}

	tomlBytes, err := toml.Marshal(data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to generate TOML",
		})
	}

	filename := strings.ReplaceAll(strings.ToLower(project.Name), " ", "_") + ".toml"
	c.Set("Content-Type", "application/toml")
	c.Set("Content-Disposition", "attachment; filename="+filename)
	return c.Send(tomlBytes)
}

// ImportProjectsRequest represents import request
type ImportProjectsRequest struct {
	Content string `json:"content"` // TOML content as string
	Mode    string `json:"mode"`    // "merge", "skip", "overwrite"
}

// ImportProjectsPreview parses TOML and returns conflicts for user decision
func ImportProjectsPreview(c fiber.Ctx) error {
	var req ImportProjectsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Parse TOML
	var data ProjectsTOML
	if err := toml.Unmarshal([]byte(req.Content), &data); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid TOML format: " + err.Error(),
		})
	}

	// Check for conflicts
	var conflicts []string
	var newProjects []string

	for _, p := range data.Projects {
		var existing database.Project
		if err := database.DB.Where("name = ?", p.Name).First(&existing).Error; err == nil {
			conflicts = append(conflicts, p.Name)
		} else {
			newProjects = append(newProjects, p.Name)
		}
	}

	return c.JSON(fiber.Map{
		"total":        len(data.Projects),
		"new_projects": newProjects,
		"conflicts":    conflicts,
		"has_conflicts": len(conflicts) > 0,
	})
}

// ImportProjectsTOML imports projects from TOML with conflict handling
func ImportProjectsTOML(c fiber.Ctx) error {
	var req ImportProjectsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Parse TOML
	var data ProjectsTOML
	if err := toml.Unmarshal([]byte(req.Content), &data); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid TOML format: " + err.Error(),
		})
	}

	mode := req.Mode
	if mode == "" {
		mode = "merge"
	}

	var imported, skipped, updated, failed int

	// Each project is imported in its own transaction so a mid-import failure
	// rolls back that project only (no half-written project/component rows)
	// instead of silently leaving the DB inconsistent.
	for _, p := range data.Projects {
		var outcome string
		err := database.DB.Transaction(func(tx *gorm.DB) error {
			o, e := importOneProject(tx, p, mode)
			outcome = o
			return e
		})
		if err != nil {
			failed++
			continue
		}
		switch outcome {
		case "imported":
			imported++
		case "updated":
			updated++
		case "skipped":
			skipped++
		}
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"message":  "Projects imported successfully",
		"imported": imported,
		"updated":  updated,
		"skipped":  skipped,
		"failed":   failed,
	})
}

// importOneProject applies a single ProjectData within transaction tx according to
// mode and returns an outcome ("imported", "updated", "skipped") or an error.
//
// Modes when a project with the same name already exists:
//   - "skip":      leave the existing project untouched
//   - "rename":    create the incoming project as a brand-new copy (unique name)
//   - "overwrite": replace the existing project's fields and components wholesale
//   - "merge":     update project fields and upsert components by name (default)
//
// A non-conflicting project is always created.
func importOneProject(tx *gorm.DB, p ProjectData, mode string) (string, error) {
	var existing database.Project
	exists := tx.Where("name = ?", p.Name).Preload("Components").First(&existing).Error == nil

	if !exists {
		return "imported", createProjectWithComponents(tx, p, p.Name)
	}

	switch mode {
	case "skip":
		return "skipped", nil

	case "rename":
		// Import as a new, independent project under an auto-suffixed name.
		return "imported", createProjectWithComponents(tx, p, uniqueProjectName(tx, p.Name))

	case "overwrite":
		tx.Where("project_id = ?", existing.ID).Delete(&database.Component{})
		applyProjectFields(&existing, p)
		if err := tx.Save(&existing).Error; err != nil {
			return "", err
		}
		for _, comp := range p.Components {
			c := buildComponent(existing.ID, comp)
			if err := tx.Create(&c).Error; err != nil {
				return "", err
			}
		}
		return "updated", nil

	default: // "merge"
		applyProjectFields(&existing, p)
		if err := tx.Save(&existing).Error; err != nil {
			return "", err
		}
		for _, comp := range p.Components {
			var ec database.Component
			if tx.Where("project_id = ? AND name = ?", existing.ID, comp.Name).First(&ec).Error == nil {
				applyComponentFields(&ec, comp)
				if err := tx.Save(&ec).Error; err != nil {
					return "", err
				}
			} else {
				c := buildComponent(existing.ID, comp)
				if err := tx.Create(&c).Error; err != nil {
					return "", err
				}
			}
		}
		return "updated", nil
	}
}

// createProjectWithComponents creates a project (under name) plus all its components.
func createProjectWithComponents(tx *gorm.DB, p ProjectData, name string) error {
	newProject := database.Project{Name: name}
	applyProjectFields(&newProject, p)
	if err := tx.Create(&newProject).Error; err != nil {
		return err
	}
	for _, comp := range p.Components {
		c := buildComponent(newProject.ID, comp)
		if err := tx.Create(&c).Error; err != nil {
			return err
		}
	}
	return nil
}

// applyProjectFields copies the non-name fields from imported data onto a project.
func applyProjectFields(dst *database.Project, p ProjectData) {
	dst.BuildServer = p.BuildServer
	dst.DeployServer = p.DeployServer
	dst.DatabaseName = p.DatabaseName
	dst.Environment = p.Environment
	dst.BackupLocation = p.BackupLocation
	dst.Description = p.Description
}

// applyComponentFields copies the non-name fields from imported data onto a component.
func applyComponentFields(dst *database.Component, comp ComponentData) {
	dst.Developer = comp.Developer
	dst.VCSType = comp.VCSType
	dst.VCSURL = comp.VCSURL
	dst.BuildCommand = comp.BuildCommand
	dst.ComponentURL = comp.ComponentURL
	dst.Enabled = comp.Enabled
	dst.Description = comp.Description
}

// buildComponent constructs a Component for projectID from imported data.
func buildComponent(projectID uint, comp ComponentData) database.Component {
	c := database.Component{ProjectID: projectID, Name: comp.Name}
	applyComponentFields(&c, comp)
	return c
}

// Helper function to merge string slices without duplicates
func mergeStringSlices(existing, incoming []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)

	for _, v := range existing {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	for _, v := range incoming {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	return result
}
