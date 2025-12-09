package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"chklst-go/internal/database"
)

// BackupManager handles database and settings backups
type BackupManager struct {
	backupDir string
}

// NewBackupManager creates a new backup manager
func NewBackupManager(backupDir string) *BackupManager {
	// Ensure backup directory exists
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		AppLogger.Error("Failed to create backup directory", err, map[string]interface{}{
			"backup_dir": backupDir,
		})
	}

	return &BackupManager{
		backupDir: backupDir,
	}
}

// BackupDatabase creates a backup of the SQLite database
func (bm *BackupManager) BackupDatabase(dbPath string) (string, error) {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupFileName := fmt.Sprintf("chklst_backup_%s.db", timestamp)
	backupPath := filepath.Join(bm.backupDir, backupFileName)

	// Open source database file
	sourceFile, err := os.Open(dbPath)
	if err != nil {
		return "", fmt.Errorf("failed to open source database: %w", err)
	}
	defer sourceFile.Close()

	// Create backup file
	destFile, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer destFile.Close()

	// Copy database file
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return "", fmt.Errorf("failed to copy database: %w", err)
	}

	AppLogger.Info("Database backed up successfully", map[string]interface{}{
		"backup_file": backupFileName,
		"backup_path": backupPath,
	})

	return backupPath, nil
}

// RestoreDatabase restores database from backup
func (bm *BackupManager) RestoreDatabase(backupPath string, targetPath string) error {
	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupPath)
	}

	// Create a backup of current database before restoring
	currentBackupName := fmt.Sprintf("pre_restore_%s.db", time.Now().Format("2006-01-02_15-04-05"))
	currentBackupPath := filepath.Join(bm.backupDir, currentBackupName)

	if _, err := os.Stat(targetPath); err == nil {
		// Current database exists, back it up first
		currentFile, err := os.Open(targetPath)
		if err == nil {
			defer currentFile.Close()
			backupFile, err := os.Create(currentBackupPath)
			if err == nil {
				defer backupFile.Close()
				io.Copy(backupFile, currentFile)
				AppLogger.Info("Current database backed up before restore", map[string]interface{}{
					"backup_file": currentBackupName,
				})
			}
		}
	}

	// Open backup file
	sourceFile, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer sourceFile.Close()

	// Create/overwrite target database
	destFile, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create target database: %w", err)
	}
	defer destFile.Close()

	// Copy backup to target
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to restore database: %w", err)
	}

	AppLogger.Info("Database restored successfully", map[string]interface{}{
		"from": backupPath,
		"to":   targetPath,
	})

	return nil
}

// ExportSettings exports library/settings to JSON
func (bm *BackupManager) ExportSettings() (string, error) {
	// Get library from database
	var library database.Library
	if err := database.DB.First(&library).Error; err != nil {
		return "", fmt.Errorf("failed to get library: %w", err)
	}

	// Create JSON file
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("settings_export_%s.json", timestamp)
	filePath := filepath.Join(bm.backupDir, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create settings file: %w", err)
	}
	defer file.Close()

	// Marshal to JSON with indentation
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(library); err != nil {
		return "", fmt.Errorf("failed to encode settings: %w", err)
	}

	AppLogger.Info("Settings exported successfully", map[string]interface{}{
		"file": fileName,
		"path": filePath,
	})

	return filePath, nil
}

// ImportSettings imports library/settings from JSON
func (bm *BackupManager) ImportSettings(filePath string) error {
	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("settings file does not exist: %s", filePath)
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open settings file: %w", err)
	}
	defer file.Close()

	// Decode JSON
	var library database.Library
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&library); err != nil {
		return fmt.Errorf("failed to decode settings: %w", err)
	}

	// Update library in database
	library.ID = 1 // Ensure we update the main library record
	if err := database.DB.Save(&library).Error; err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	AppLogger.Info("Settings imported successfully", map[string]interface{}{
		"from": filePath,
	})

	return nil
}

// ListBackups lists all backup files
func (bm *BackupManager) ListBackups() ([]string, error) {
	entries, err := os.ReadDir(bm.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []string
	for _, entry := range entries {
		if !entry.IsDir() && (filepath.Ext(entry.Name()) == ".db" || filepath.Ext(entry.Name()) == ".json") {
			backups = append(backups, entry.Name())
		}
	}

	return backups, nil
}

// CleanOldBackups removes backups older than specified days
func (bm *BackupManager) CleanOldBackups(daysToKeep int) error {
	entries, err := os.ReadDir(bm.backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	cutoffTime := time.Now().AddDate(0, 0, -daysToKeep)
	removed := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoffTime) {
			filePath := filepath.Join(bm.backupDir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				AppLogger.Error("Failed to remove old backup", err, map[string]interface{}{
					"file": entry.Name(),
				})
			} else {
				removed++
			}
		}
	}

	AppLogger.Info("Cleaned old backups", map[string]interface{}{
		"removed": removed,
		"older_than_days": daysToKeep,
	})

	return nil
}

// StartAutoBackup starts automatic daily backups
func (bm *BackupManager) StartAutoBackup(dbPath string, intervalHours int) {
	ticker := time.NewTicker(time.Duration(intervalHours) * time.Hour)

	go func() {
		for range ticker.C {
			_, err := bm.BackupDatabase(dbPath)
			if err != nil {
				AppLogger.Error("Auto-backup failed", err, nil)
			}

			// Clean backups older than 30 days
			if err := bm.CleanOldBackups(30); err != nil {
				AppLogger.Error("Failed to clean old backups", err, nil)
			}
		}
	}()

	AppLogger.Info("Auto-backup started", map[string]interface{}{
		"interval_hours": intervalHours,
	})
}
