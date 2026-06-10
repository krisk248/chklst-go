package handlers

import (
	"chklst-go/internal/ai"
	"chklst-go/internal/database"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

// defaultOllamaURL resolves the Ollama base URL, preferring the OLLAMA_URL env
// var (set in Docker to reach the host) and falling back to localhost.
func defaultOllamaURL() string {
	if v := os.Getenv("OLLAMA_URL"); v != "" {
		return v
	}
	return ai.DefaultBaseURL
}

// envDefault sets *field from env var `key` when *field is blank. Returns true if changed.
func envDefault(field *string, key string) bool {
	if *field == "" {
		if v := os.Getenv(key); v != "" {
			*field = v
			return true
		}
	}
	return false
}

// envDefaultStr is like envDefault but uses a literal fallback when the env var is unset.
func envDefaultStr(field *string, key, fallback string) bool {
	if *field == "" {
		if v := os.Getenv(key); v != "" {
			*field = v
		} else {
			*field = fallback
		}
		return true
	}
	return false
}

// applyAIDefaults backfills AI + email settings on rows created before these
// columns existed, so the UI always sees usable defaults.
func applyAIDefaults(s *database.Settings) bool {
	changed := false
	if s.OllamaURL == "" {
		s.OllamaURL = defaultOllamaURL()
		changed = true
	}
	if s.AIModel == "" {
		s.AIModel = ai.DefaultModel
		changed = true
	}
	// Pre-fill mail settings from env (only blanks; never overrides saved values).
	// Keeping personal/site values out of the code so the repo carries no secrets.
	changed = envDefault(&s.SMTPHost, "SMTP_HOST") || changed
	changed = envDefaultStr(&s.SMTPSecurity, "SMTP_SECURITY", "starttls") || changed
	changed = envDefault(&s.SMTPUser, "SMTP_USER") || changed
	changed = envDefault(&s.SMTPFrom, "SMTP_FROM") || changed
	changed = envDefault(&s.SMTPTo, "SMTP_TO") || changed
	changed = envDefault(&s.SMTPCc, "SMTP_CC") || changed
	changed = envDefault(&s.SMTPTestTo, "SMTP_TEST_TO") || changed
	if s.SMTPPort == 0 {
		if p, err := strconv.Atoi(os.Getenv("SMTP_PORT")); err == nil && p > 0 {
			s.SMTPPort = p
		} else {
			s.SMTPPort = 587
		}
		changed = true
	}
	if s.SummaryGenerateTime == "" {
		s.SummaryGenerateTime = "18:30"
		changed = true
	}
	if s.SummarySendTime == "" {
		s.SummarySendTime = "20:00"
		changed = true
	}
	if s.SummaryWeekdays == "" {
		s.SummaryWeekdays = "0,1,2,3,4" // Sunday..Thursday
		changed = true
	}
	return changed
}

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
			OllamaURL:          defaultOllamaURL(),
			AIModel:            ai.DefaultModel,
		}
		if err := database.DB.Create(&settings).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to create default settings",
			})
		}
	} else if applyAIDefaults(&settings) {
		database.DB.Save(&settings)
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

	// Update existing settings - basic fields
	settings.DefaultDeployedBy = req.DefaultDeployedBy
	settings.ExcelExportPath = req.ExcelExportPath
	settings.AutoClearAfterSave = req.AutoClearAfterSave

	// Update Jira fields
	settings.JiraURL = req.JiraURL
	settings.JiraEmail = req.JiraEmail
	settings.JiraToken = req.JiraToken
	settings.JiraProject = req.JiraProject

	// Update webhook fields
	settings.TeamsWebhookURL = req.TeamsWebhookURL
	settings.WebhooksEnabled = req.WebhooksEnabled

	// Update AI / Ollama fields
	settings.AIEnabled = req.AIEnabled
	settings.OllamaURL = req.OllamaURL
	settings.AIModel = req.AIModel
	settings.PlannedWeekly = req.PlannedWeekly
	settings.ParsonSystemPrompt = req.ParsonSystemPrompt
	settings.ParsonTemperature = req.ParsonTemperature

	// Git Insights (GitHub) fields
	settings.GitHubToken = req.GitHubToken
	settings.GitHubOwner = req.GitHubOwner
	settings.GitHubRepos = req.GitHubRepos

	// Email (SMTP) fields
	settings.SMTPHost = req.SMTPHost
	settings.SMTPPort = req.SMTPPort
	settings.SMTPUser = req.SMTPUser
	settings.SMTPPassword = req.SMTPPassword
	settings.SMTPFrom = req.SMTPFrom
	settings.SMTPTo = req.SMTPTo
	settings.SMTPCc = req.SMTPCc
	settings.SMTPSecurity = req.SMTPSecurity
	settings.SMTPTestTo = req.SMTPTestTo

	// Scheduler fields
	settings.SummaryScheduleEnabled = req.SummaryScheduleEnabled
	settings.SummaryGenerateTime = req.SummaryGenerateTime
	settings.SummarySendTime = req.SummarySendTime
	settings.SummaryWeekdays = req.SummaryWeekdays
	settings.SummaryAutoSend = req.SummaryAutoSend

	applyAIDefaults(&settings)

	if err := database.DB.Save(&settings).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update settings",
		})
	}

	return c.JSON(settings)
}
