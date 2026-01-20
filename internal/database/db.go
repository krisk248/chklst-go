package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase initializes the SQLite database connection
func InitDatabase(dbPath string) error {
	var err error

	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Info)

	// Open SQLite connection
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger:                 gormLogger,
		SkipDefaultTransaction: true, // Better performance
		PrepareStmt:            true, // Cache prepared statements
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("✅ Database connection established")

	return nil
}

// AutoMigrate runs auto-migration for all models
func AutoMigrate() error {
	log.Println("🔄 Running auto-migration...")

	err := DB.AutoMigrate(
		&Project{},
		&Component{},
		&Deployment{},
		&Library{},
		&Settings{},
	)

	if err != nil {
		return fmt.Errorf("auto-migration failed: %w", err)
	}

	log.Println("✅ Auto-migration completed")

	// Ensure default library exists
	var library Library
	result := DB.First(&library)
	if result.Error == gorm.ErrRecordNotFound {
		library = Library{
			ID:            1,
			Developers:    StringArray{"Kannan"},
			BuildServers:  StringArray{"192.168.1.149"},
			DeployServers: StringArray{},
			Environments:  StringArray{"QA", "UAT", "Production"},
		}
		if err := DB.Create(&library).Error; err != nil {
			log.Printf("⚠️  Warning: Failed to create default library: %v", err)
		} else {
			log.Println("✅ Default library created")
		}
	}

	return nil
}

// CloseDatabase closes the database connection
func CloseDatabase() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
