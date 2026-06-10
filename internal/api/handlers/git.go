package handlers

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"chklst-go/internal/database"
	"chklst-go/internal/git"

	"github.com/gofiber/fiber/v3"
)

// gitClient builds a GitHub client from settings.
func gitClient() (*git.Client, database.Settings, error) {
	var s database.Settings
	database.DB.First(&s)
	if strings.TrimSpace(s.GitHubOwner) == "" {
		return nil, s, fmt.Errorf("set a GitHub owner in Settings")
	}
	return git.NewClient(s.GitHubToken, s.GitHubOwner), s, nil
}

// TestGitHub verifies the token + owner.
func TestGitHub(c fiber.Ctx) error {
	client, _, err := gitClient()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"ok": false, "error": err.Error()})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := client.TestConnection(ctx); err != nil {
		return c.Status(400).JSON(fiber.Map{"ok": false, "error": err.Error()})
	}
	repos, _ := client.ListRepos(ctx)
	return c.JSON(fiber.Map{"ok": true, "owner": client.Owner, "repos_found": len(repos)})
}

// --- response types ---

type authorStat struct {
	Name        string `json:"name"`
	Commits     int    `json:"commits"`
	Repos       int    `json:"repos"`
	Merges      int    `json:"merges"`
	LastActive  string `json:"last_active"`
}

type dayCount struct {
	Label  string `json:"label"`
	Date   string `json:"date"`
	Count  int    `json:"count"`
}

type repoStat struct {
	Name    string `json:"name"`
	Commits int    `json:"commits"`
	Merges  int    `json:"merges"`
}

// GitInsightsResponse is the deterministic commit-metric payload.
type GitInsightsResponse struct {
	PeriodDays   int          `json:"period_days"`
	TotalCommits int          `json:"total_commits"`
	TotalMerges  int          `json:"total_merges"` // proxy for integration/CI cadence
	Repos        int          `json:"repos"`
	Authors      []authorStat `json:"authors"`     // most active first
	ByDay        []dayCount   `json:"by_day"`       // last N days
	ByRepo       []repoStat   `json:"by_repo"`
	ByHour       [24]int      `json:"by_hour"`      // commit cadence (local hour)
}

// GetGitInsights fetches commits across the configured repos and computes metrics.
// Query: ?days=N (default 30).
func GetGitInsights(c fiber.Ctx) error {
	client, s, err := gitClient()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	days, _ := strconv.Atoi(c.Query("days"))
	if days <= 0 || days > 180 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Resolve repo list.
	repos := splitCSV(s.GitHubRepos)
	if len(repos) == 0 {
		repos, err = client.ListRepos(ctx)
		if err != nil {
			return c.Status(502).JSON(fiber.Map{"error": "Could not list repos: " + err.Error()})
		}
	}

	var all []git.Commit
	for _, r := range repos {
		commits, err := client.CommitsSince(ctx, r, since)
		if err != nil {
			// Skip a repo we can't read, but keep going.
			continue
		}
		all = append(all, commits...)
	}

	return c.JSON(buildGitInsights(all, repos, days))
}

func buildGitInsights(commits []git.Commit, repos []string, days int) GitInsightsResponse {
	resp := GitInsightsResponse{PeriodDays: days, Repos: len(repos), TotalCommits: len(commits)}

	type aAcc struct {
		commits, merges int
		repos           map[string]bool
		last            time.Time
	}
	byAuthor := map[string]*aAcc{}
	byRepo := map[string]*repoStat{}
	byDay := map[string]int{}

	for _, c := range commits {
		// author
		a := byAuthor[c.Author]
		if a == nil {
			a = &aAcc{repos: map[string]bool{}}
			byAuthor[c.Author] = a
		}
		a.commits++
		a.repos[c.Repo] = true
		if c.IsMerge {
			a.merges++
			resp.TotalMerges++
		}
		if c.Date.After(a.last) {
			a.last = c.Date
		}
		// repo
		r := byRepo[c.Repo]
		if r == nil {
			r = &repoStat{Name: c.Repo}
			byRepo[c.Repo] = r
		}
		r.Commits++
		if c.IsMerge {
			r.Merges++
		}
		// day + hour (local)
		lt := c.Date.In(time.Local)
		byDay[lt.Format("2006-01-02")]++
		resp.ByHour[lt.Hour()]++
	}

	for name, a := range byAuthor {
		resp.Authors = append(resp.Authors, authorStat{
			Name: name, Commits: a.commits, Repos: len(a.repos), Merges: a.merges,
			LastActive: a.last.In(time.Local).Format("Jan 2"),
		})
	}
	sort.Slice(resp.Authors, func(i, j int) bool { return resp.Authors[i].Commits > resp.Authors[j].Commits })

	for _, r := range byRepo {
		resp.ByRepo = append(resp.ByRepo, *r)
	}
	sort.Slice(resp.ByRepo, func(i, j int) bool { return resp.ByRepo[i].Commits > resp.ByRepo[j].Commits })

	// Last `days` days as an ordered series.
	now := time.Now()
	limit := days
	if limit > 30 {
		limit = 30 // graph stays readable; metrics above still cover the full period
	}
	for i := limit - 1; i >= 0; i-- {
		d := now.AddDate(0, 0, -i)
		key := d.Format("2006-01-02")
		resp.ByDay = append(resp.ByDay, dayCount{Label: d.Format("Jan 2"), Date: key, Count: byDay[key]})
	}
	return resp
}

func splitCSV(s string) []string {
	out := []string{}
	for _, p := range strings.Split(s, ",") {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
