package database

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// StringArray is a custom type for JSON string arrays in SQLite
type StringArray []string

// Scan implements sql.Scanner interface
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("failed to unmarshal StringArray value")
	}

	return json.Unmarshal(bytes, a)
}

// Value implements driver.Valuer interface
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "[]", nil
	}
	return json.Marshal(a)
}

// Project represents a deployment project
type Project struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	Name           string       `gorm:"unique;not null;index" json:"name"`
	BuildServer    string       `json:"build_server"`
	DeployServer   string       `json:"deploy_server"`
	DatabaseName   string       `json:"database_name"`
	Environment    string       `json:"environment"`
	BackupLocation string       `json:"backup_location"`
	Description    string       `gorm:"type:text" json:"description"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`

	// Relationships
	Components  []Component  `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"components,omitempty"`
	Deployments []Deployment `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"deployments,omitempty"`
}

// Component represents a project component
type Component struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ProjectID    uint      `gorm:"not null;index" json:"project_id"`
	Name         string    `gorm:"not null;index" json:"name"`
	Developer    string    `json:"developer"`
	VCSType      string    `gorm:"default:'git'" json:"vcs_type"` // git, svn, etc.
	VCSURL       string    `json:"vcs_url"`
	BuildCommand string    `json:"build_command"`
	ComponentURL string    `json:"component_url"`
	Enabled      bool      `gorm:"default:true" json:"enabled"`
	Description  string    `gorm:"type:text" json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Project     Project      `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Deployments []Deployment `gorm:"foreignKey:ComponentID;constraint:OnDelete:CASCADE" json:"deployments,omitempty"`
}

// Deployment represents a deployment record
type Deployment struct {
	ID                   uint       `gorm:"primaryKey" json:"id"`
	JiraID               string     `gorm:"index" json:"jira_id"`
	Timestamp            time.Time  `gorm:"index" json:"timestamp"`
	ProjectID            uint       `gorm:"not null;index" json:"project_id"`
	ComponentID          *uint      `gorm:"index" json:"component_id"` // Nullable for legacy data
	Environment          string     `json:"environment"`
	VCSURL               string     `json:"vcs_url"`
	DeveloperName        string     `json:"developer_name"`
	BuildServer          string     `json:"build_server"`
	DeployServer         string     `json:"deploy_server"`
	DatabaseName         string     `json:"database_name"`
	DBBackupLocation     string     `json:"db_backup_location"`
	DatabaseScript       string     `gorm:"type:text" json:"database_script"`
	PreviousBuildBackup  string     `json:"previous_build_backup"`
	BuildStatus          string     `gorm:"default:'pending'" json:"build_status"`
	DeployStatus         string     `gorm:"default:'pending'" json:"deploy_status"`
	Notes                string     `gorm:"type:text" json:"notes"`
	DeployedBy           string     `json:"deployed_by"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`

	// Relationships
	Project   Project    `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Component *Component `gorm:"foreignKey:ComponentID" json:"component,omitempty"`
}

// Library stores preset values for dropdowns
type Library struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	Developers    StringArray `gorm:"type:json" json:"developers"`
	BuildServers  StringArray `gorm:"type:json" json:"build_servers"`
	DeployServers StringArray `gorm:"type:json" json:"deploy_servers"`
	Environments  StringArray `gorm:"type:json" json:"environments"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// TableName overrides the table name for Library
func (Library) TableName() string {
	return "library"
}

// Settings stores application settings (singleton)
type Settings struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	DefaultDeployedBy  string    `json:"default_deployed_by"`
	ExcelExportPath    string    `json:"excel_export_path"`
	AutoClearAfterSave bool      `json:"auto_clear_after_save"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// TableName overrides the table name for Settings
func (Settings) TableName() string {
	return "settings"
}
