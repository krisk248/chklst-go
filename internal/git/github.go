// Package git fetches commit history from GitHub for the "Git Insights" analytics.
package git

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const apiBase = "https://api.github.com"

// Client talks to the GitHub REST API.
type Client struct {
	Token string
	Owner string
	http  *http.Client
}

func NewClient(token, owner string) *Client {
	return &Client{Token: token, Owner: owner, http: &http.Client{Timeout: 30 * time.Second}}
}

// Commit is one commit, flattened from the GitHub response.
type Commit struct {
	SHA     string    `json:"sha"`
	Repo    string    `json:"repo"`
	Author  string    `json:"author"`  // GitHub login if available, else commit author name
	Email   string    `json:"email"`
	Date    time.Time `json:"date"`
	Message string    `json:"message"`
	IsMerge bool      `json:"is_merge"`
}

func (c *Client) do(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBase+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("github request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("github auth failed (check token)")
	}
	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("github rate limit or access forbidden")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github returned status %d for %s", resp.StatusCode, path)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// ListRepos returns repo names owned by Owner (most-recently-pushed first, up to 100).
func (c *Client) ListRepos(ctx context.Context) ([]string, error) {
	var repos []struct {
		Name string `json:"name"`
	}
	// /users/{owner}/repos works for both users and orgs for public + accessible repos.
	if err := c.do(ctx, fmt.Sprintf("/users/%s/repos?per_page=100&sort=pushed", url.PathEscape(c.Owner)), &repos); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(repos))
	for _, r := range repos {
		out = append(out, r.Name)
	}
	return out, nil
}

// ghCommit mirrors the relevant part of the GitHub commits response.
type ghCommit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Name  string    `json:"name"`
			Email string    `json:"email"`
			Date  time.Time `json:"date"`
		} `json:"author"`
	} `json:"commit"`
	Author  *struct {
		Login string `json:"login"`
	} `json:"author"`
	Parents []struct {
		SHA string `json:"sha"`
	} `json:"parents"`
}

// CommitsSince fetches commits for one repo since `since` (paginated, capped at 5 pages
// = 500 commits per repo to bound cost).
func (c *Client) CommitsSince(ctx context.Context, repo string, since time.Time) ([]Commit, error) {
	var out []Commit
	for page := 1; page <= 5; page++ {
		var batch []ghCommit
		path := fmt.Sprintf("/repos/%s/%s/commits?since=%s&per_page=100&page=%d",
			url.PathEscape(c.Owner), url.PathEscape(repo), since.UTC().Format(time.RFC3339), page)
		if err := c.do(ctx, path, &batch); err != nil {
			return out, err
		}
		for _, gc := range batch {
			author := gc.Commit.Author.Name
			if gc.Author != nil && gc.Author.Login != "" {
				author = gc.Author.Login
			}
			out = append(out, Commit{
				SHA:     gc.SHA,
				Repo:    repo,
				Author:  strings.TrimSpace(author),
				Email:   gc.Commit.Author.Email,
				Date:    gc.Commit.Author.Date,
				Message: strings.SplitN(gc.Commit.Message, "\n", 2)[0],
				IsMerge: len(gc.Parents) > 1,
			})
		}
		if len(batch) < 100 {
			break
		}
	}
	return out, nil
}

// TestConnection verifies the token works (GET /user or the owner's repos).
func (c *Client) TestConnection(ctx context.Context) error {
	var who struct {
		Login string `json:"login"`
	}
	if err := c.do(ctx, "/user", &who); err != nil {
		// Fall back to a public read of the owner so a fine-grained token still validates.
		_, rerr := c.ListRepos(ctx)
		return rerr
	}
	return nil
}
