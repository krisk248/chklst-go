package handlers

import (
	"chklst-go/internal/utils"
	"os"

	"github.com/gofiber/fiber/v3"
)

var backupManager *utils.BackupManager

// InitAdminHandlers initializes the admin handlers
func InitAdminHandlers(bm *utils.BackupManager) {
	backupManager = bm
}

// BackupDatabase creates a database backup
func BackupDatabase(c fiber.Ctx) error {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./chklst.db"
	}

	backupPath, err := backupManager.BackupDatabase(dbPath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to backup database",
		})
	}

	return c.JSON(fiber.Map{
		"message":     "Database backed up successfully",
		"backup_path": backupPath,
	})
}

// RestoreDatabase restores database from backup
func RestoreDatabase(c fiber.Ctx) error {
	var req struct {
		BackupPath string `json:"backup_path"`
	}

	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./chklst.db"
	}

	if err := backupManager.RestoreDatabase(req.BackupPath, dbPath); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to restore database",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Database restored successfully",
	})
}

// ExportSettings exports library settings to JSON
func ExportSettings(c fiber.Ctx) error {
	filePath, err := backupManager.ExportSettings()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to export settings",
		})
	}

	return c.JSON(fiber.Map{
		"message":   "Settings exported successfully",
		"file_path": filePath,
	})
}

// ImportSettings imports library settings from JSON
func ImportSettings(c fiber.Ctx) error {
	var req struct {
		FilePath string `json:"file_path"`
	}

	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := backupManager.ImportSettings(req.FilePath); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to import settings",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Settings imported successfully",
	})
}

// ListBackups lists all available backups
func ListBackups(c fiber.Ctx) error {
	backups, err := backupManager.ListBackups()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to list backups",
		})
	}

	return c.JSON(fiber.Map{
		"backups": backups,
	})
}
