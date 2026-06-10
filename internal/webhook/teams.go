package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TeamsClient handles Microsoft Teams webhook notifications
type TeamsClient struct {
	WebhookURL string
	HTTPClient *http.Client
}

// NewTeamsClient creates a new Teams webhook client
func NewTeamsClient(webhookURL string) *TeamsClient {
	return &TeamsClient{
		WebhookURL: webhookURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// DeploymentInfo contains deployment data for notifications
type DeploymentInfo struct {
	JiraID       string
	ProjectName  string
	Component    string
	Environment  string
	Developer    string
	DeployedBy   string
	BuildStatus  string
	DeployStatus string
	Timestamp    time.Time
}

// TeamsMessage represents a simple Teams message
type TeamsMessage struct {
	Type       string       `json:"@type"`
	Context    string       `json:"@context"`
	ThemeColor string       `json:"themeColor"`
	Summary    string       `json:"summary"`
	Sections   []Section    `json:"sections"`
	Actions    []CardAction `json:"potentialAction,omitempty"`
}

// Section represents a message section
type Section struct {
	ActivityTitle    string `json:"activityTitle,omitempty"`
	ActivitySubtitle string `json:"activitySubtitle,omitempty"`
	ActivityImage    string `json:"activityImage,omitempty"`
	Facts            []Fact `json:"facts,omitempty"`
	Markdown         bool   `json:"markdown"`
}

// Fact represents a key-value fact in Teams message
type Fact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// CardAction represents an action button
type CardAction struct {
	Type    string   `json:"@type"`
	Name    string   `json:"name"`
	Targets []Target `json:"targets,omitempty"`
}

// Target represents action target URL
type Target struct {
	OS  string `json:"os"`
	URI string `json:"uri"`
}

// SendDeploymentNotification sends a deployment notification to Teams
func (c *TeamsClient) SendDeploymentNotification(info DeploymentInfo) error {
	// Choose color based on status
	color := "0076D7" // blue for pending
	if info.DeployStatus == "success" {
		color = "2DC72D" // green
	} else if info.DeployStatus == "failed" {
		color = "FF0000" // red
	}

	statusEmoji := "🚀"
	if info.DeployStatus == "success" {
		statusEmoji = "✅"
	} else if info.DeployStatus == "failed" {
		statusEmoji = "❌"
	}

	msg := TeamsMessage{
		Type:       "MessageCard",
		Context:    "http://schema.org/extensions",
		ThemeColor: color,
		Summary:    fmt.Sprintf("Deployment: %s - %s", info.ProjectName, info.DeployStatus),
		Sections: []Section{
			{
				ActivityTitle:    fmt.Sprintf("%s Deployment: %s", statusEmoji, info.JiraID),
				ActivitySubtitle: fmt.Sprintf("Deployed by %s at %s", info.DeployedBy, info.Timestamp.Format("2006-01-02 15:04")),
				Markdown:         true,
				Facts: []Fact{
					{Name: "Project", Value: info.ProjectName},
					{Name: "Component", Value: info.Component},
					{Name: "Environment", Value: info.Environment},
					{Name: "Developer", Value: info.Developer},
					{Name: "Build Status", Value: info.BuildStatus},
					{Name: "Deploy Status", Value: info.DeployStatus},
				},
			},
		},
	}

	return c.Send(msg)
}

// Send sends a raw Teams message
func (c *TeamsClient) Send(msg TeamsMessage) error {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshaling message: %w", err)
	}

	resp, err := c.HTTPClient.Post(c.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("sending webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("teams webhook error: status %d", resp.StatusCode)
	}

	return nil
}

// SendSimpleMessage sends a plain text message
func (c *TeamsClient) SendSimpleMessage(text string) error {
	msg := map[string]string{"text": text}
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshaling message: %w", err)
	}

	resp, err := c.HTTPClient.Post(c.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("sending webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("teams webhook error: status %d", resp.StatusCode)
	}

	return nil
}

// TestWebhook sends a test message to verify the webhook works
func (c *TeamsClient) TestWebhook() error {
	return c.SendSimpleMessage("🔔 Test notification from chklst-go - webhook configured successfully!")
}
