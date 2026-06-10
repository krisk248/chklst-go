package jira

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client handles Jira API interactions
type Client struct {
	BaseURL    string
	Email      string
	Token      string
	Project    string
	HTTPClient *http.Client
}

// NewClient creates a new Jira client
func NewClient(baseURL, email, token, project string) *Client {
	return &Client{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		Email:   email,
		Token:   token,
		Project: project,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Issue represents a Jira issue
type Issue struct {
	ID     string      `json:"id"`
	Key    string      `json:"key"`
	Self   string      `json:"self"`
	Fields IssueFields `json:"fields"`
}

// IssueFields contains the issue field data
type IssueFields struct {
	Summary     string       `json:"summary"`
	Description *Description `json:"description"`
	Status      *Status      `json:"status"`
	Assignee    *User        `json:"assignee"`
	Reporter    *User        `json:"reporter"`
	Created     string       `json:"created"`
	Updated     string       `json:"updated"`
	Priority    *Priority    `json:"priority"`
	IssueType   *IssueType   `json:"issuetype"`

	// Custom fields for deployment
	BuildServer  string         `json:"customfield_10178"` // short text
	DeployServer string         `json:"customfield_10179"` // short text
	Database     string         `json:"customfield_10181"` // short text
	DBBackup     string         `json:"customfield_10180"` // short text
	Environment  *CustomOption  `json:"customfield_10176"` // dropdown
	Notes        string         `json:"customfield_10177"` // paragraph
	Developers   []User         `json:"customfield_10174"` // multi-user picker
}

// Description represents Atlassian Document Format content
type Description struct {
	Type    string    `json:"type"`
	Version int       `json:"version"`
	Content []Content `json:"content"`
}

// Content represents ADF content blocks
type Content struct {
	Type    string    `json:"type"`
	Content []Text    `json:"content,omitempty"`
	Text    string    `json:"text,omitempty"`
}

// Text represents text content in ADF
type Text struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Status represents issue status
type Status struct {
	Name string `json:"name"`
}

// User represents a Jira user
type User struct {
	AccountID   string `json:"accountId"`
	DisplayName string `json:"displayName"`
	Email       string `json:"emailAddress"`
}

// Priority represents issue priority
type Priority struct {
	Name string `json:"name"`
}

// IssueType represents the issue type
type IssueType struct {
	Name string `json:"name"`
}

// CustomOption represents a dropdown option
type CustomOption struct {
	Value string `json:"value"`
}

// SearchResult represents JQL search response
type SearchResult struct {
	Total      int     `json:"total"`
	MaxResults int     `json:"maxResults"`
	StartAt    int     `json:"startAt"`
	Issues     []Issue `json:"issues"`
}

// authHeader returns the Basic Auth header value
func (c *Client) authHeader() string {
	auth := fmt.Sprintf("%s:%s", c.Email, c.Token)
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

// doRequest executes an HTTP request with auth
func (c *Client) doRequest(method, endpoint string, body io.Reader) ([]byte, error) {
	url := c.BaseURL + endpoint

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("jira API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetMyIssues fetches issues assigned to the authenticated user
func (c *Client) GetMyIssues() ([]Issue, error) {
	jql := fmt.Sprintf("project = %s AND assignee = currentUser() ORDER BY updated DESC", c.Project)
	return c.SearchIssues(jql)
}

// GetTodayIssues fetches issues updated today
func (c *Client) GetTodayIssues() ([]Issue, error) {
	jql := fmt.Sprintf("project = %s AND assignee = currentUser() AND updated >= startOfDay() ORDER BY updated DESC", c.Project)
	return c.SearchIssues(jql)
}

// GetIssuesByStatus fetches issues by status
func (c *Client) GetIssuesByStatus(status string) ([]Issue, error) {
	jql := fmt.Sprintf("project = %s AND assignee = currentUser() AND status = \"%s\" ORDER BY updated DESC", c.Project, status)
	return c.SearchIssues(jql)
}

// SearchIssues executes a JQL search using the new Jira API (POST /rest/api/3/search/jql)
func (c *Client) SearchIssues(jql string) ([]Issue, error) {
	// New Jira API uses POST with JSON body
	body := map[string]interface{}{
		"jql":        jql,
		"maxResults": 50,
		"fields":     []string{"*all"},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling search request: %w", err)
	}

	data, err := c.doRequest("POST", "/rest/api/3/search/jql", strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, err
	}

	var result SearchResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing search results: %w", err)
	}

	return result.Issues, nil
}

// GetIssue fetches a single issue by key
func (c *Client) GetIssue(key string) (*Issue, error) {
	endpoint := fmt.Sprintf("/rest/api/3/issue/%s?fields=*all", key)

	data, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var issue Issue
	if err := json.Unmarshal(data, &issue); err != nil {
		return nil, fmt.Errorf("parsing issue: %w", err)
	}

	return &issue, nil
}

// AddComment adds a comment to an issue
func (c *Client) AddComment(issueKey, comment string) error {
	endpoint := fmt.Sprintf("/rest/api/3/issue/%s/comment", issueKey)

	// Atlassian Document Format (ADF) for the comment
	body := map[string]interface{}{
		"body": map[string]interface{}{
			"type":    "doc",
			"version": 1,
			"content": []map[string]interface{}{
				{
					"type": "paragraph",
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": comment,
						},
					},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling comment: %w", err)
	}

	_, err = c.doRequest("POST", endpoint, strings.NewReader(string(jsonBody)))
	return err
}

// TestConnection verifies the Jira credentials work
func (c *Client) TestConnection() error {
	_, err := c.doRequest("GET", "/rest/api/3/myself", nil)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	return nil
}

// GetDescriptionText extracts plain text from ADF description
func (issue *Issue) GetDescriptionText() string {
	if issue.Fields.Description == nil {
		return ""
	}

	var text strings.Builder
	for _, content := range issue.Fields.Description.Content {
		if content.Type == "paragraph" {
			for _, t := range content.Content {
				text.WriteString(t.Text)
			}
			text.WriteString("\n")
		}
	}
	return strings.TrimSpace(text.String())
}

// GetEnvironmentValue returns the environment dropdown value
func (issue *Issue) GetEnvironmentValue() string {
	if issue.Fields.Environment == nil {
		return ""
	}
	return issue.Fields.Environment.Value
}

// GetDeveloperNames returns comma-separated developer names
func (issue *Issue) GetDeveloperNames() string {
	if len(issue.Fields.Developers) == 0 {
		return ""
	}
	var names []string
	for _, dev := range issue.Fields.Developers {
		names = append(names, dev.DisplayName)
	}
	return strings.Join(names, ", ")
}
