package handlers

import (
	"chklst-go/internal/database"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
)

// ListDeployments returns all deployments with optional filtering
func ListDeployments(c fiber.Ctx) error {
	query := database.DB.Preload("Project").Preload("Component")

	// Filter by project
	if projectID := c.Query("project_id"); projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	// Filter by month/year
	if month := c.Query("month"); month != "" {
		if year := c.Query("year"); year != "" {
			monthInt, _ := strconv.Atoi(month)
			yearInt, _ := strconv.Atoi(year)
			
			startDate := time.Date(yearInt, time.Month(monthInt), 1, 0, 0, 0, 0, time.UTC)
			endDate := startDate.AddDate(0, 1, 0)
			
			query = query.Where("timestamp >= ? AND timestamp < ?", startDate, endDate)
		}
	}

	var deployments []database.Deployment
	if err := query.Find(&deployments).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch deployments",
		})
	}

	return c.JSON(deployments)
}

// GetDeployment returns a single deployment by ID
func GetDeployment(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid deployment ID",
		})
	}

	var deployment database.Deployment
	if err := database.DB.Preload("Project").Preload("Component").First(&deployment, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Deployment not found",
		})
	}

	return c.JSON(deployment)
}

// DeploymentRequest is used for creating deployments with flexible timestamp parsing
type DeploymentRequest struct {
	JiraID             string  `json:"jira_id"`
	ProjectID          uint    `json:"project_id"`
	ComponentID        *uint   `json:"component_id"`
	Timestamp          string  `json:"timestamp"` // Accept as string for flexible parsing
	Environment        string  `json:"environment"`
	VCSURL             string  `json:"vcs_url"`
	DeveloperName      string  `json:"developer_name"`
	BuildServer        string  `json:"build_server"`
	DeployServer       string  `json:"deploy_server"`
	DatabaseName       string  `json:"database_name"`
	DBBackupLocation   string  `json:"db_backup_location"`
	DatabaseScript     string  `json:"database_script"`
	PreviousBuildBackup string `json:"previous_build_backup"`
	BuildStatus        string  `json:"build_status"`
	DeployStatus       string  `json:"deploy_status"`
	Notes              string  `json:"notes"`
	DeployedBy         string  `json:"deployed_by"`
}

// parseTimestamp flexibly parses timestamp in multiple formats
func parseTimestamp(ts string) (time.Time, error) {
	if ts == "" {
		return time.Now(), nil
	}

	// Try RFC3339 format (with timezone): "2025-12-09T16:23:43Z"
	if t, err := time.Parse(time.RFC3339, ts); err == nil {
		return t, nil
	}

	// Try datetime-local format (Vue sends): "2025-12-09T16:23:43"
	if t, err := time.Parse("2006-01-02T15:04:05", ts); err == nil {
		return t, nil
	}

	// Try datetime-local WITHOUT seconds: "2025-12-09T16:39"
	if t, err := time.Parse("2006-01-02T15:04", ts); err == nil {
		return t, nil
	}

	// Try with added Z (convert to UTC)
	if t, err := time.Parse(time.RFC3339, ts+"Z"); err == nil {
		return t, nil
	}

	// Try ISO8601 with milliseconds
	if t, err := time.Parse("2006-01-02T15:04:05.999Z07:00", ts); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("invalid timestamp format: %s", ts)
}

// CreateDeployment creates a new deployment
func CreateDeployment(c fiber.Ctx) error {
	var req DeploymentRequest

	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Parse timestamp flexibly
	timestamp, err := parseTimestamp(req.Timestamp)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid timestamp: %s", err.Error()),
		})
	}

	// Create deployment from request
	deployment := database.Deployment{
		JiraID:              req.JiraID,
		ProjectID:           req.ProjectID,
		ComponentID:         req.ComponentID,
		Timestamp:           timestamp,
		Environment:         req.Environment,
		VCSURL:              req.VCSURL,
		DeveloperName:       req.DeveloperName,
		BuildServer:         req.BuildServer,
		DeployServer:        req.DeployServer,
		DatabaseName:        req.DatabaseName,
		DBBackupLocation:    req.DBBackupLocation,
		DatabaseScript:      req.DatabaseScript,
		PreviousBuildBackup: req.PreviousBuildBackup,
		BuildStatus:         req.BuildStatus,
		DeployStatus:        req.DeployStatus,
		Notes:               req.Notes,
		DeployedBy:          req.DeployedBy,
	}

	if err := database.DB.Create(&deployment).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create deployment",
		})
	}

	// Preload relationships
	database.DB.Preload("Project").Preload("Component").First(&deployment, deployment.ID)

	return c.Status(201).JSON(deployment)
}

// UpdateDeployment updates an existing deployment
func UpdateDeployment(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid deployment ID",
		})
	}

	var deployment database.Deployment
	if err := database.DB.First(&deployment, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Deployment not found",
		})
	}

	if err := c.Bind().JSON(&deployment); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := database.DB.Save(&deployment).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update deployment",
		})
	}

	return c.JSON(deployment)
}

// DeleteDeployment deletes a deployment
func DeleteDeployment(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid deployment ID",
		})
	}

	if err := database.DB.Delete(&database.Deployment{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete deployment",
		})
	}

	return c.SendStatus(204)
}
