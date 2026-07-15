package handlers

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"chklst-go/internal/ai"
	"chklst-go/internal/database"

	"github.com/gofiber/fiber/v3"
)

// maxNoteLen bounds free-text inputs before they reach the model — a basic
// guardrail against oversized prompts and prompt-injection padding.
const maxNoteLen = 4000

// genMu serializes report generation across the HTTP handler and the scheduler.
// CPU inference takes ~1-2 min; without this, overlapping runs would race and the
// later writer would silently overwrite the earlier result.
var genMu sync.Mutex

// summaryDate returns the target date string ("2006-01-02"), defaulting to today
// in the server's local timezone (set via TZ in Docker).
func summaryDate(c fiber.Ctx) string {
	if d := c.Query("date"); d != "" {
		return d
	}
	return time.Now().Format("2006-01-02")
}

// deploymentsForDate loads deployments whose LOCAL date matches dateStr, with
// project + component preloaded. A coarse DB window keeps the query cheap; the
// exact date match is done in Go so it's correct regardless of how the timestamp
// timezone is stored.
func deploymentsForDate(dateStr string) ([]database.Deployment, error) {
	target, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		return nil, err
	}
	start := target.AddDate(0, 0, -1)
	end := target.AddDate(0, 0, 2)

	var deps []database.Deployment
	if err := database.DB.Preload("Project").Preload("Component").
		Where("timestamp >= ? AND timestamp < ?", start, end).
		Order("timestamp asc").Find(&deps).Error; err != nil {
		return nil, err
	}

	out := make([]database.Deployment, 0, len(deps))
	for _, d := range deps {
		if d.Timestamp.In(time.Local).Format("2006-01-02") == dateStr {
			out = append(out, d)
		}
	}
	return out, nil
}

// summaryStats are deterministic, rule-based facts computed in Go (not by the AI),
// surfaced both to the UI and to the prompt.
type summaryStats struct {
	Total          int            `json:"total"`
	ByEnvironment  map[string]int `json:"by_environment"`
	DirtyPatches   []string       `json:"dirty_patches"`    // deployments with no JIRA ID
	HighChurn      []string       `json:"high_churn"`       // projects with 2+ patches today
	FailedCount    int            `json:"failed_count"`
}

func computeStats(deps []database.Deployment) summaryStats {
	// Initialize slices so empty results marshal to [] (not null), which would
	// break `.length` checks in the frontend.
	s := summaryStats{ByEnvironment: map[string]int{}, DirtyPatches: []string{}, HighChurn: []string{}}
	s.Total = len(deps)
	perProject := map[string]int{}
	for _, d := range deps {
		s.ByEnvironment[orUnknown(d.Environment)]++
		perProject[projectLabel(d)]++
		if strings.TrimSpace(d.JiraID) == "" {
			s.DirtyPatches = append(s.DirtyPatches, projectLabel(d)+" ("+orUnknown(d.Environment)+")")
		}
		if d.DeployStatus == "failed" || d.BuildStatus == "failed" {
			s.FailedCount++
		}
	}
	// High churn: a project deployed 2+ times today (per the bad-patch rule).
	for proj, n := range perProject {
		if n >= 2 {
			s.HighChurn = append(s.HighChurn, fmt.Sprintf("%s (%d patches)", proj, n))
		}
	}
	sort.Strings(s.HighChurn)
	return s
}

func orUnknown(v string) string {
	if strings.TrimSpace(v) == "" {
		return "Unknown"
	}
	return v
}

func projectLabel(d database.Deployment) string {
	if d.Project.Name != "" {
		return d.Project.Name
	}
	return fmt.Sprintf("Project %d", d.ProjectID)
}

// getOrCreateSummary returns the DailySummary row for a date, creating a draft if absent.
func getOrCreateSummary(date string) (database.DailySummary, error) {
	var summary database.DailySummary
	err := database.DB.Where("date = ?", date).First(&summary).Error
	if err != nil {
		summary = database.DailySummary{Date: date, Status: "draft"}
		if cerr := database.DB.Create(&summary).Error; cerr != nil {
			return summary, cerr
		}
	}
	return summary, nil
}

// GetTodaySummary returns the draft for the target date plus the deployments and
// computed stats that will feed the report.
func GetTodaySummary(c fiber.Ctx) error {
	date := summaryDate(c)

	summary, err := getOrCreateSummary(date)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load summary"})
	}

	deps, err := deploymentsForDate(date)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load deployments: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"summary":     summary,
		"deployments": deps,
		"stats":       computeStats(deps),
	})
}

// ListRecentSummaries returns the latest summaries for the Parson activity tracker.
func ListRecentSummaries(c fiber.Ctx) error {
	var rows []database.DailySummary
	if err := database.DB.Order("date desc").Limit(60).Find(&rows).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load activity"})
	}
	return c.JSON(rows)
}

// UpdateSummary persists the human-entered notes and any manual edits to the AI output.
func UpdateSummary(c fiber.Ctx) error {
	id := c.Params("id")
	var summary database.DailySummary
	if err := database.DB.First(&summary, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Summary not found"})
	}

	var req database.DailySummary
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Detect a real edit to an already-generated report (review activity).
	editedAfterGen := summary.Status == "generated" &&
		(req.GeneratedBody != summary.GeneratedBody || req.GeneratedSubject != summary.GeneratedSubject)

	summary.Notes = clip(req.Notes)
	summary.Planned = clip(req.Planned)
	summary.Roadblocks = clip(req.Roadblocks)
	summary.GeneratedSubject = req.GeneratedSubject
	summary.GeneratedBody = req.GeneratedBody
	summary.Recipients = req.Recipients

	if editedAfterGen {
		now := time.Now()
		summary.Reviewed = true
		summary.LockedAt = &now // re-lock after the edit
	}

	if err := database.DB.Save(&summary).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save summary"})
	}
	return c.JSON(summary)
}

// GenerateSummary builds an augmented prompt from real deployment data + notes and
// asks the local model to write the report. The facts are gathered in Go; the model
// only formats and narrates, so it cannot invent deployments.
//
// Optional body field {"model": "gemma4:12b"} overrides the configured model for this
// generation only — used by the UI to compare models side by side.
func GenerateSummary(c fiber.Ctx) error {
	id := c.Params("id")
	var summary database.DailySummary
	if err := database.DB.First(&summary, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Summary not found"})
	}

	var settings database.Settings
	database.DB.First(&settings)
	if !settings.AIEnabled {
		return c.Status(400).JSON(fiber.Map{"error": "AI is disabled. Enable it in Settings > AI."})
	}

	var genReq struct {
		Model string `json:"model"`
	}
	_ = c.Bind().JSON(&genReq)

	if !genMu.TryLock() {
		return c.Status(409).JSON(fiber.Map{"error": "A generation is already running — wait for it to finish."})
	}
	elapsed, model, err := generateSummaryRow(&summary, settings, genReq.Model, false)
	genMu.Unlock()
	if err != nil {
		return c.Status(503).JSON(fiber.Map{"error": "AI generation failed: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"summary":            summary,
		"generation_seconds": elapsed.Seconds(),
		"model":              model,
	})
}

// generateSummaryRow runs Parson over the summary's day and stores the result.
// Shared by the HTTP handler and the scheduler. modelOverride (optional) overrides
// the configured model for this run; auto marks scheduler-produced reports.
func generateSummaryRow(summary *database.DailySummary, settings database.Settings, modelOverride string, auto bool) (time.Duration, string, error) {
	deps, err := deploymentsForDate(summary.Date)
	if err != nil {
		return 0, "", err
	}

	prompt := buildSummaryPrompt(*summary, deps, computeStats(deps), settings.PlannedWeekly)

	client := aiClientFromSettings()
	if modelOverride != "" {
		client.Model = modelOverride
	}
	// Cold model load can take minutes on CPU; give it generous room.
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()

	// Agent config: custom system prompt / temperature if set, else defaults.
	systemPrompt := summarySystemPrompt
	if strings.TrimSpace(settings.ParsonSystemPrompt) != "" {
		systemPrompt = settings.ParsonSystemPrompt
	}
	temperature := defaultTemperature
	if settings.ParsonTemperature > 0 {
		temperature = settings.ParsonTemperature
	}

	started := time.Now()
	out, err := client.Generate(ctx, prompt, ai.GenerateOptions{
		System:      systemPrompt,
		Temperature: temperature,
	})
	if err != nil {
		return 0, client.Model, err
	}
	elapsed := time.Since(started)

	out = stripThinking(out)
	subject, body := splitSubject(out, summary.Date)
	if strings.TrimSpace(body) == "" {
		return elapsed, client.Model, errors.New("model returned an empty summary")
	}

	now := time.Now()
	summary.GeneratedSubject = subject
	summary.GeneratedBody = body
	summary.Status = "generated"
	summary.GeneratedAt = &now
	summary.GenerationSeconds = elapsed.Seconds()
	summary.GeneratedByModel = client.Model
	summary.LockedAt = &now // generated = locked (until edited)
	summary.Reviewed = false
	summary.AutoGenerated = auto
	if err := database.DB.Save(summary).Error; err != nil {
		return elapsed, client.Model, err
	}
	return elapsed, client.Model, nil
}

// defaultTemperature is Parson's default sampling temperature for the daily report —
// low for consistent formatting.
const defaultTemperature = 0.4

const summarySystemPrompt = `You write a daily deployment status email for a DevOps engineer to send to their team.
Use ONLY the facts the user provides. Never invent deployments, ticket numbers, projects, components, environments, dates, or notes.
Be concise and professional. Correct any spelling and grammar mistakes in the user's notes, but keep their meaning.
No greeting, no sign-off, no preamble. Output ONLY the email, in plain text (no markdown, no bold, no asterisks).

Use EXACTLY this structure:

Subject: Daily Activity Report – <Month D, YYYY>

Today's Activities:
- Patch Deployments Completed (<N> Total) — <ENVIRONMENT>
   - <PAT-NUMBER>
      - <Project> (<Component>)
      - Deployed to <ENVIRONMENT>
      - Notes: <notes>
      - Time: <time>
- <each extra activity from notes as its own bullet, if any>

Planned Activities:
- <each planned item as a bullet>

Road Blocks / Suggestions:
- <each roadblock as a bullet>

Rules:
- The subject date must use the full month name, e.g. "June 8, 2026".
- Under "Patch Deployments Completed (N Total) — ENV": N is the count of deployments for that environment. If patches span multiple environments, repeat the header for each environment.
- Each deployment is given as: jira | project | component | environment | status | time | notes.
- Write one block per deployment: the PAT number as its own indented line, then beneath it four bullet lines in this exact order: project (component), "Deployed to <ENV>", "Notes: <notes>", "Time: <time>".
- If the same PAT number was deployed more than once, write a separate block for each deployment, repeating the PAT number line.
- If a deployment's notes value is "none", leave the Notes line out entirely. ALWAYS include the Time line.
- Copy the time exactly as given (e.g. "5:42 PM"). Never write a date on deployment lines — only the time.
- If a patch has no component, write just the project name without parentheses.
- For deployments whose jira value is "(no JIRA)", the block's first line reads exactly: No JIRA ID
- If there are NO patch deployments, write under Today's Activities exactly: "- No patches deployed today." Then if the notes describe other work, add those as bullets.
- If no planned activities are given, output exactly: "- Nothing specified."
- If no roadblocks are given, output exactly: "- No roadblocks."

Example (format reference only — do NOT reuse this data):
Subject: Daily Activity Report – April 21, 2026

Today's Activities:
- Patch Deployments Completed (3 Total) — QA
   - PAT-652
      - MBANK (MBANK-BE)
      - Deployed to QA
      - Notes: Payment gateway config updated
      - Time: 10:15 AM
   - PAT-652
      - MBANK (MBANK-FE)
      - Deployed to QA
      - Time: 10:40 AM
   - No JIRA ID
      - ADX-IPO (ADXIPO)
      - Deployed to QA
      - Notes: Hotfix for login page
      - Time: 2:30 PM

Planned Activities:
- Nothing specified.

Road Blocks / Suggestions:
- No roadblocks.`

// niceDate turns "2006-01-02" into "January 2, 2006" (full month name).
func niceDate(dateStr string) string {
	t, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		return dateStr
	}
	return t.Format("January 2, 2006")
}

// buildSummaryPrompt renders the real data into the user-side prompt. weeklyPlanned
// is the standing "planned activities for the week" text (from Settings), combined
// with the day-specific planned activities.
func buildSummaryPrompt(s database.DailySummary, deps []database.Deployment, stats summaryStats, weeklyPlanned string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "DATE: %s\n\n", niceDate(s.Date))

	fmt.Fprintf(&b, "PATCH DEPLOYMENTS (%d) (jira | project | component | environment | status | time | notes):\n", len(deps))
	if len(deps) == 0 {
		b.WriteString("- none\n")
	}
	for _, d := range deps {
		jira := d.JiraID
		if jira == "" {
			jira = "(no JIRA)"
		}
		comp := ""
		if d.Component != nil && d.Component.Name != "" {
			comp = d.Component.Name
		}
		fmt.Fprintf(&b, "- %s | %s | %s | %s | %s | %s | %s\n",
			jira, projectLabel(d), orDash(comp), orUnknown(d.Environment), orUnknown(d.DeployStatus),
			d.Timestamp.In(time.Local).Format("3:04 PM"),
			orNone(clip(d.Notes)))
	}

	// Planned = standing weekly plan + day-specific plan.
	planned := strings.TrimSpace(strings.TrimSpace(weeklyPlanned) + "\n" + strings.TrimSpace(s.Planned))

	fmt.Fprintf(&b, "\nEXTRA ACTIVITIES (from notes): %s\n", orNone(s.Notes))
	fmt.Fprintf(&b, "PLANNED ACTIVITIES: %s\n", orNone(planned))
	fmt.Fprintf(&b, "ROADBLOCKS: %s\n", orNone(s.Roadblocks))

	if stats.FailedCount > 0 {
		fmt.Fprintf(&b, "\nNOTE: %d failed deployment(s) today — mention under Road Blocks.\n", stats.FailedCount)
	}
	return b.String()
}

func orDash(v string) string {
	if strings.TrimSpace(v) == "" {
		return "—"
	}
	return v
}

func orNone(v string) string {
	if strings.TrimSpace(v) == "" {
		return "none"
	}
	return v
}

func clip(v string) string {
	v = strings.TrimSpace(v)
	if len(v) > maxNoteLen {
		return v[:maxNoteLen]
	}
	return v
}

// splitSubject extracts the "Subject:" line; the remainder is the body. Falls back
// to a default subject if the model didn't emit one.
func splitSubject(out, date string) (subject, body string) {
	lines := strings.Split(out, "\n")
	for i, ln := range lines {
		if strings.HasPrefix(strings.TrimSpace(ln), "Subject:") {
			subject = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(ln), "Subject:"))
			body = strings.TrimSpace(strings.Join(append(lines[:i], lines[i+1:]...), "\n"))
			return subject, body
		}
	}
	return "Daily Activity Report – " + niceDate(date), strings.TrimSpace(out)
}

// stripThinking removes reasoning blocks some models (e.g. Gemma 4 thinking mode)
// may emit, so only the final report is stored.
func stripThinking(out string) string {
	for _, open := range []string{"<think>", "<thinking>"} {
		close := strings.Replace(open, "<", "</", 1)
		for {
			i := strings.Index(out, open)
			if i < 0 {
				break
			}
			j := strings.Index(out, close)
			if j < 0 || j < i {
				out = strings.Replace(out, open, "", 1)
				continue
			}
			out = out[:i] + out[j+len(close):]
		}
	}
	return strings.TrimSpace(out)
}
