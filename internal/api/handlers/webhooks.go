package handlers

import (
	"chklst-go/internal/database"
	"chklst-go/internal/webhook"

	"github.com/gofiber/fiber/v3"
)

// TestTeamsWebhook tests the Teams webhook configuration
func TestTeamsWebhook(c fiber.Ctx) error {
	var settings database.Settings
	if err := database.DB.First(&settings).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Settings not found",
		})
	}

	if settings.TeamsWebhookURL == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Teams webhook URL not configured",
		})
	}

	if !settings.WebhooksEnabled {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Webhooks are disabled",
		})
	}

	client := webhook.NewTeamsClient(settings.TeamsWebhookURL)
	if err := client.TestWebhook(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Webhook test failed: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Test notification sent successfully",
	})
}

// SendDeploymentWebhook sends a deployment notification
// This is typically called after a deployment is saved
func SendDeploymentWebhook(deployment database.Deployment) error {
	var settings database.Settings
	if err := database.DB.First(&settings).Error; err != nil {
		return err
	}

	if !settings.WebhooksEnabled || settings.TeamsWebhookURL == "" {
		return nil // Silently skip if not configured
	}

	// Get project name
	var project database.Project
	database.DB.First(&project, deployment.ProjectID)

	// Get component name if available
	componentName := ""
	if deployment.ComponentID != nil {
		var component database.Component
		if database.DB.First(&component, *deployment.ComponentID).Error == nil {
			componentName = component.Name
		}
	}

	client := webhook.NewTeamsClient(settings.TeamsWebhookURL)
	info := webhook.DeploymentInfo{
		JiraID:       deployment.JiraID,
		ProjectName:  project.Name,
		Component:    componentName,
		Environment:  deployment.Environment,
		Developer:    deployment.DeveloperName,
		DeployedBy:   deployment.DeployedBy,
		BuildStatus:  deployment.BuildStatus,
		DeployStatus: deployment.DeployStatus,
		Timestamp:    deployment.Timestamp,
	}

	return client.SendDeploymentNotification(info)
}
