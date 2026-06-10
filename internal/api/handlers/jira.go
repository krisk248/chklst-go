package handlers

import (
	"chklst-go/internal/database"
	"chklst-go/internal/jira"

	"github.com/gofiber/fiber/v3"
)

// getJiraClient creates a Jira client from settings
func getJiraClient() (*jira.Client, error) {
	var settings database.Settings
	if err := database.DB.First(&settings).Error; err != nil {
		return nil, err
	}

	if settings.JiraURL == "" || settings.JiraEmail == "" || settings.JiraToken == "" {
		return nil, fiber.NewError(fiber.StatusBadRequest, "Jira not configured")
	}

	return jira.NewClient(settings.JiraURL, settings.JiraEmail, settings.JiraToken, settings.JiraProject), nil
}

// GetJiraTickets returns tickets assigned to the user
func GetJiraTickets(c fiber.Ctx) error {
	client, err := getJiraClient()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Optional filters
	status := c.Query("status")
	todayOnly := c.Query("today") == "true"
	var issues []jira.Issue

	if status != "" {
		issues, err = client.GetIssuesByStatus(status)
	} else if todayOnly {
		issues, err = client.GetTodayIssues()
	} else {
		issues, err = client.GetMyIssues()
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch Jira tickets: " + err.Error(),
		})
	}

	// Transform to simpler format for frontend
	tickets := make([]map[string]interface{}, len(issues))
	for i, issue := range issues {
		tickets[i] = map[string]interface{}{
			"key":           issue.Key,
			"summary":       issue.Fields.Summary,
			"status":        getStatusName(issue.Fields.Status),
			"priority":      getPriorityName(issue.Fields.Priority),
			"assignee":      getDisplayName(issue.Fields.Assignee),
			"reporter":      getDisplayName(issue.Fields.Reporter),
			"created":       issue.Fields.Created,
			"updated":       issue.Fields.Updated,
			"description":   issue.GetDescriptionText(),
			"build_server":  issue.Fields.BuildServer,
			"deploy_server": issue.Fields.DeployServer,
			"database":      issue.Fields.Database,
			"db_backup":     issue.Fields.DBBackup,
			"environment":   issue.GetEnvironmentValue(),
			"notes":         issue.Fields.Notes,
			"developers":    issue.GetDeveloperNames(),
		}
	}

	return c.JSON(fiber.Map{
		"tickets": tickets,
		"total":   len(tickets),
	})
}

// GetJiraTicket returns a single Jira ticket by key
func GetJiraTicket(c fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Ticket key is required",
		})
	}

	client, err := getJiraClient()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	issue, err := client.GetIssue(key)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch ticket: " + err.Error(),
		})
	}

	ticket := map[string]interface{}{
		"key":           issue.Key,
		"summary":       issue.Fields.Summary,
		"status":        getStatusName(issue.Fields.Status),
		"priority":      getPriorityName(issue.Fields.Priority),
		"assignee":      getDisplayName(issue.Fields.Assignee),
		"reporter":      getDisplayName(issue.Fields.Reporter),
		"created":       issue.Fields.Created,
		"updated":       issue.Fields.Updated,
		"description":   issue.GetDescriptionText(),
		"build_server":  issue.Fields.BuildServer,
		"deploy_server": issue.Fields.DeployServer,
		"database":      issue.Fields.Database,
		"db_backup":     issue.Fields.DBBackup,
		"environment":   issue.GetEnvironmentValue(),
		"notes":         issue.Fields.Notes,
		"developers":    issue.GetDeveloperNames(),
	}

	return c.JSON(ticket)
}

// PostJiraComment adds a comment to a Jira ticket
type CommentRequest struct {
	Comment string `json:"comment"`
}

func PostJiraComment(c fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Ticket key is required",
		})
	}

	var req CommentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Comment == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Comment is required",
		})
	}

	client, err := getJiraClient()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if err := client.AddComment(key, req.Comment); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to add comment: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Comment added successfully",
	})
}

// TestJiraConnection tests the Jira API connection
func TestJiraConnection(c fiber.Ctx) error {
	client, err := getJiraClient()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	if err := client.TestConnection(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Connection test failed: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Jira connection successful",
	})
}

// Helper functions
func getStatusName(status *jira.Status) string {
	if status == nil {
		return ""
	}
	return status.Name
}

func getPriorityName(priority *jira.Priority) string {
	if priority == nil {
		return ""
	}
	return priority.Name
}

func getDisplayName(user *jira.User) string {
	if user == nil {
		return ""
	}
	return user.DisplayName
}
