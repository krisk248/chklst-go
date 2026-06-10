package handlers

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"chklst-go/internal/ai"
	"chklst-go/internal/database"

	"github.com/gofiber/fiber/v3"
)

// ProjectInsight is the per-project health computed deterministically from deployments.
type ProjectInsight struct {
	ProjectID       uint     `json:"project_id"`
	Name            string   `json:"name"`
	Deployments     int      `json:"deployments"`
	SuccessRate     int      `json:"success_rate"`     // %
	JiraCompliance  int      `json:"jira_compliance"`  // %
	DirtyPatches    int      `json:"dirty_patches"`    // deployments with no JIRA
	HighChurnDays   int      `json:"high_churn_days"`  // days with 2+ patches (RED)
	OverFreqWeeks   int      `json:"over_freq_weeks"`  // weeks with >3 patches (too frequent)
	Failed          int      `json:"failed"`
	BusiestDay      string   `json:"busiest_day"`      // date with most patches
	BusiestDayCount int      `json:"busiest_day_count"`
	HealthScore     int      `json:"health_score"`     // 0..100
	Classification  string   `json:"classification"`   // good | attention | dirty
	Flags           []string `json:"flags"`
}

// SummaryCounts is a quick rollup for a timeframe (this week / this month).
type SummaryCounts struct {
	Deployments  int `json:"deployments"`
	Projects     int `json:"projects"`
	DirtyPatches int `json:"dirty_patches"` // no-JIRA
	RedDays      int `json:"red_days"`      // days with 2+ patches
}

// WeekCount is one bar in the frequency graph.
type WeekCount struct {
	Label string `json:"label"` // e.g. "Jun 9"
	Count int    `json:"count"`
}

// DeveloperStat is per-developer discipline within the selected period.
type DeveloperStat struct {
	Name           string `json:"name"`
	Deployments    int    `json:"deployments"`
	SuccessRate    int    `json:"success_rate"`
	JiraCompliance int    `json:"jira_compliance"`
	Projects       int    `json:"projects"`
	Failed         int    `json:"failed"`
}

// EnvStat is per-environment risk within the selected period.
type EnvStat struct {
	Name        string `json:"name"`
	Deployments int    `json:"deployments"`
	SuccessRate int    `json:"success_rate"`
	Failed      int    `json:"failed"`
	NoJira      int    `json:"no_jira"`
}

// MonthTrend is one point of the month-over-month health trend (always the last
// 6 calendar months, independent of the period filter).
type MonthTrend struct {
	Label          string `json:"label"` // e.g. "Jan 26"
	Deployments    int    `json:"deployments"`
	SuccessRate    int    `json:"success_rate"`
	JiraCompliance int    `json:"jira_compliance"`
	RedDays        int    `json:"red_days"`
}

// InsightsResponse is the analytics payload for a period.
type InsightsResponse struct {
	Period         string           `json:"period"`
	TotalDeploys   int              `json:"total_deployments"`
	TotalProjects  int              `json:"total_projects"`
	SuccessRate    int              `json:"success_rate"`
	JiraCompliance int              `json:"jira_compliance"`
	DirtyPatches   int              `json:"dirty_patches"`
	Projects       []ProjectInsight `json:"projects"` // sorted worst-health first

	// Period-scoped breakdowns (respect the month/year filter).
	Developers   []DeveloperStat `json:"developers"`
	Environments []EnvStat       `json:"environments"`

	// Always current (independent of the period filter), for the top summary + graph.
	ThisWeek        SummaryCounts `json:"this_week"`
	ThisMonth       SummaryCounts `json:"this_month"`
	WeeklyFrequency []WeekCount   `json:"weekly_frequency"` // last 10 weeks
	MonthlyTrends   []MonthTrend  `json:"monthly_trends"`   // last 6 months
}

// deploymentsForPeriod loads deployments filtered by optional month (1-12) and year,
// pushing the date range into SQL so large histories stay cheap. month==0 means any
// month; year==0 means any year. Project is preloaded.
func deploymentsForPeriod(month, year int) ([]database.Deployment, error) {
	q := database.DB.Preload("Project").Order("timestamp asc")

	switch {
	case month != 0 && year != 0:
		start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
		q = q.Where("timestamp >= ? AND timestamp < ?", start, start.AddDate(0, 1, 0))
	case year != 0:
		start := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
		q = q.Where("timestamp >= ? AND timestamp < ?", start, start.AddDate(1, 0, 0))
	case month != 0:
		// Month across all years is rare; fall back to in-Go filtering below.
	}

	var deps []database.Deployment
	if err := q.Find(&deps).Error; err != nil {
		return nil, err
	}
	if month != 0 && year == 0 {
		out := make([]database.Deployment, 0, len(deps))
		for _, d := range deps {
			if int(d.Timestamp.In(time.Local).Month()) == month {
				out = append(out, d)
			}
		}
		return out, nil
	}
	return deps, nil
}

// GetInsights computes per-project health insights for a period.
// Query: ?month=1..12 (optional) &year=YYYY (optional). Defaults to all time.
func GetInsights(c fiber.Ctx) error {
	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))

	periodDeps, err := deploymentsForPeriod(month, year)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load deployments"})
	}
	recentDeps, err := deploymentsSince(time.Now().AddDate(0, -6, 0))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load recent deployments"})
	}

	resp := buildInsights(periodDeps, month, year)
	addTimeframe(&resp, recentDeps)
	return c.JSON(resp)
}

// deploymentsSince loads deployments from t onward (used for the rolling
// this-week/this-month/frequency rollups without scanning the whole table).
func deploymentsSince(t time.Time) ([]database.Deployment, error) {
	var deps []database.Deployment
	err := database.DB.Where("timestamp >= ?", t).Order("timestamp asc").Find(&deps).Error
	return deps, err
}

// addTimeframe fills the always-current this-week / this-month rollups, the
// last-10-weeks frequency series, and the last-6-months trend. allDeps should
// cover at least the last 6 months.
func addTimeframe(resp *InsightsResponse, allDeps []database.Deployment) {
	now := time.Now()
	thisWeek := weekKey(now)
	resp.MonthlyTrends = computeMonthlyTrends(allDeps, now)

	var weekDeps, monthDeps []database.Deployment
	for _, d := range allDeps {
		t := d.Timestamp.In(time.Local)
		if weekKey(t) == thisWeek {
			weekDeps = append(weekDeps, d)
		}
		if t.Year() == now.Year() && t.Month() == now.Month() {
			monthDeps = append(monthDeps, d)
		}
	}
	resp.ThisWeek = computeSummary(weekDeps)
	resp.ThisMonth = computeSummary(monthDeps)

	// Frequency: counts per ISO week, last 10 weeks ending this week.
	counts := map[string]int{}
	for _, d := range allDeps {
		counts[weekKey(d.Timestamp.In(time.Local))]++
	}
	for i := 9; i >= 0; i-- {
		t := now.AddDate(0, 0, -7*i)
		resp.WeeklyFrequency = append(resp.WeeklyFrequency, WeekCount{
			Label: mondayOf(t).Format("Jan 2"),
			Count: counts[weekKey(t)],
		})
	}
}

// computeSummary rolls up a slice of deployments (deployments, distinct projects,
// no-JIRA patches, and red days = (project,day) pairs with 2+ patches).
func computeSummary(deps []database.Deployment) SummaryCounts {
	s := SummaryCounts{Deployments: len(deps)}
	projs := map[uint]bool{}
	perPD := map[string]int{}
	for _, d := range deps {
		projs[d.ProjectID] = true
		if strings.TrimSpace(d.JiraID) == "" {
			s.DirtyPatches++
		}
		perPD[fmt.Sprintf("%d|%s", d.ProjectID, d.Timestamp.In(time.Local).Format("2006-01-02"))]++
	}
	s.Projects = len(projs)
	for _, n := range perPD {
		if n >= 2 {
			s.RedDays++
		}
	}
	return s
}

// mondayOf returns the Monday of t's week (for week labels).
func mondayOf(t time.Time) time.Time {
	wd := int(t.Weekday())
	if wd == 0 {
		wd = 7 // Sunday -> 7 so Monday is start
	}
	return t.AddDate(0, 0, -(wd - 1))
}

// buildInsights does the deterministic analytics. Exposed (lowercase) for reuse by
// the AI narrative endpoint.
func buildInsights(deps []database.Deployment, month, year int) InsightsResponse {
	type acc struct {
		name     string
		total    int
		success  int
		withJira int
		dirty    int
		failed   int
		perDay   map[string]int
		perWeek  map[string]int
	}
	byProject := map[uint]*acc{}

	totalSuccess, totalJira := 0, 0
	for _, d := range deps {
		a := byProject[d.ProjectID]
		if a == nil {
			a = &acc{name: projectLabel(d), perDay: map[string]int{}, perWeek: map[string]int{}}
			byProject[d.ProjectID] = a
		}
		a.total++
		if d.DeployStatus == "success" {
			a.success++
			totalSuccess++
		}
		if strings.TrimSpace(d.JiraID) != "" {
			a.withJira++
			totalJira++
		} else {
			a.dirty++
		}
		if d.DeployStatus == "failed" || d.BuildStatus == "failed" {
			a.failed++
		}
		t := d.Timestamp.In(time.Local)
		a.perDay[t.Format("2006-01-02")]++
		a.perWeek[weekKey(t)]++
	}

	projects := make([]ProjectInsight, 0, len(byProject))
	for pid, a := range byProject {
		pi := ProjectInsight{ProjectID: pid, Name: a.name, Deployments: a.total, Failed: a.failed, DirtyPatches: a.dirty}
		if a.total > 0 {
			pi.SuccessRate = a.success * 100 / a.total
			pi.JiraCompliance = a.withJira * 100 / a.total
		}
		for day, n := range a.perDay {
			if n >= 2 {
				pi.HighChurnDays++ // RED: 2+ patches in a single day
			}
			if n > pi.BusiestDayCount {
				pi.BusiestDayCount = n
				pi.BusiestDay = day
			}
		}
		for _, n := range a.perWeek {
			if n > 3 {
				pi.OverFreqWeeks++ // too frequent: >3 patches in a week
			}
		}
		pi.HealthScore = healthScore(pi)
		pi.Classification, pi.Flags = classify(pi)
		projects = append(projects, pi)
	}

	// Worst health first, so problems surface at the top.
	sort.Slice(projects, func(i, j int) bool {
		if projects[i].HealthScore != projects[j].HealthScore {
			return projects[i].HealthScore < projects[j].HealthScore
		}
		return projects[i].Deployments > projects[j].Deployments
	})

	resp := InsightsResponse{
		Period:        periodLabel(month, year),
		TotalDeploys:  len(deps),
		TotalProjects: len(byProject),
		DirtyPatches:  countDirty(deps),
		Projects:      projects,
		Developers:    computeDeveloperStats(deps),
		Environments:  computeEnvStats(deps),
	}
	if len(deps) > 0 {
		resp.SuccessRate = totalSuccess * 100 / len(deps)
		resp.JiraCompliance = totalJira * 100 / len(deps)
	}
	return resp
}

// computeDeveloperStats aggregates per-developer discipline (most active first).
func computeDeveloperStats(deps []database.Deployment) []DeveloperStat {
	type acc struct {
		total, success, jira, failed int
		projects                     map[uint]bool
	}
	byDev := map[string]*acc{}
	for _, d := range deps {
		name := strings.TrimSpace(d.DeveloperName)
		if name == "" {
			name = "(unassigned)"
		}
		a := byDev[name]
		if a == nil {
			a = &acc{projects: map[uint]bool{}}
			byDev[name] = a
		}
		a.total++
		a.projects[d.ProjectID] = true
		if d.DeployStatus == "success" {
			a.success++
		}
		if strings.TrimSpace(d.JiraID) != "" {
			a.jira++
		}
		if d.DeployStatus == "failed" || d.BuildStatus == "failed" {
			a.failed++
		}
	}
	out := make([]DeveloperStat, 0, len(byDev))
	for name, a := range byDev {
		out = append(out, DeveloperStat{
			Name:           name,
			Deployments:    a.total,
			SuccessRate:    a.success * 100 / a.total,
			JiraCompliance: a.jira * 100 / a.total,
			Projects:       len(a.projects),
			Failed:         a.failed,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Deployments > out[j].Deployments })
	return out
}

// computeEnvStats aggregates per-environment risk (most active first).
func computeEnvStats(deps []database.Deployment) []EnvStat {
	type acc struct{ total, success, failed, noJira int }
	byEnv := map[string]*acc{}
	for _, d := range deps {
		env := orUnknown(d.Environment)
		a := byEnv[env]
		if a == nil {
			a = &acc{}
			byEnv[env] = a
		}
		a.total++
		if d.DeployStatus == "success" {
			a.success++
		}
		if d.DeployStatus == "failed" || d.BuildStatus == "failed" {
			a.failed++
		}
		if strings.TrimSpace(d.JiraID) == "" {
			a.noJira++
		}
	}
	out := make([]EnvStat, 0, len(byEnv))
	for env, a := range byEnv {
		out = append(out, EnvStat{
			Name:        env,
			Deployments: a.total,
			SuccessRate: a.success * 100 / a.total,
			Failed:      a.failed,
			NoJira:      a.noJira,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Deployments > out[j].Deployments })
	return out
}

// computeMonthlyTrends builds the last-6-months month-over-month trend.
func computeMonthlyTrends(deps []database.Deployment, now time.Time) []MonthTrend {
	type acc struct {
		total, success, jira int
		perPD                map[string]int
	}
	byMonth := map[string]*acc{}
	for _, d := range deps {
		t := d.Timestamp.In(time.Local)
		key := t.Format("2006-01")
		a := byMonth[key]
		if a == nil {
			a = &acc{perPD: map[string]int{}}
			byMonth[key] = a
		}
		a.total++
		if d.DeployStatus == "success" {
			a.success++
		}
		if strings.TrimSpace(d.JiraID) != "" {
			a.jira++
		}
		a.perPD[fmt.Sprintf("%d|%s", d.ProjectID, t.Format("2006-01-02"))]++
	}

	out := make([]MonthTrend, 0, 6)
	for i := 5; i >= 0; i-- {
		m := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local).AddDate(0, -i, 0)
		key := m.Format("2006-01")
		mt := MonthTrend{Label: m.Format("Jan 06")}
		if a := byMonth[key]; a != nil {
			mt.Deployments = a.total
			mt.SuccessRate = a.success * 100 / a.total
			mt.JiraCompliance = a.jira * 100 / a.total
			for _, n := range a.perPD {
				if n >= 2 {
					mt.RedDays++
				}
			}
		}
		out = append(out, mt)
	}
	return out
}

// healthScore starts at 100 and subtracts CAPPED penalties per the user's rules:
// no-JIRA patches and 2+/day "red" days hurt the most, >3/week frequency and
// failures hurt moderately. Caps keep the score a usable gradient even over long
// windows (so the worst projects sink but cleaner ones still rank higher).
func healthScore(p ProjectInsight) int {
	if p.Deployments == 0 {
		return 0
	}
	score := 100
	score -= minInt(30, p.DirtyPatches*5) // no-JIRA: definite flag
	score -= minInt(35, p.HighChurnDays*4) // 2+/day: RED flag (heaviest)
	score -= minInt(15, p.OverFreqWeeks*3) // >3/week: too frequent
	score -= minInt(20, p.Failed*10)       // failures
	if score < 0 {
		score = 0
	}
	return score
}

// classify gives a score-based health gradient, while the flags always spell out the
// specific rule violations (🔴 no-JIRA / 2+ a day / failures, ⚠ over-frequency).
func classify(p ProjectInsight) (string, []string) {
	flags := []string{} // never nil — a nil slice marshals to JSON null and breaks p.flags.length
	if p.DirtyPatches > 0 {
		flags = append(flags, "🔴 "+strconv.Itoa(p.DirtyPatches)+" no-JIRA patch(es)")
	}
	if p.HighChurnDays > 0 {
		flags = append(flags, "🔴 "+strconv.Itoa(p.HighChurnDays)+" day(s) with 2+ patches")
	}
	if p.OverFreqWeeks > 0 {
		flags = append(flags, "⚠ "+strconv.Itoa(p.OverFreqWeeks)+" week(s) over 3 patches")
	}
	if p.Failed > 0 {
		flags = append(flags, "🔴 "+strconv.Itoa(p.Failed)+" failed deployment(s)")
	}

	switch {
	case p.HealthScore >= 80:
		return "good", flags
	case p.HealthScore >= 50:
		return "attention", flags
	default:
		return "dirty", flags
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// weekKey returns a stable ISO year-week key, e.g. "2026-W24".
func weekKey(t time.Time) string {
	y, w := t.ISOWeek()
	return fmt.Sprintf("%d-W%02d", y, w)
}

func countDirty(deps []database.Deployment) int {
	n := 0
	for _, d := range deps {
		if strings.TrimSpace(d.JiraID) == "" {
			n++
		}
	}
	return n
}

const analyticsSystemPrompt = `You are Parson, a DevOps deployment analyst. You are given pre-computed, factual deployment health metrics. Use ONLY these facts — never invent projects, numbers, or dates.

Write a concise analysis with these sections (plain text, short bullets):

This Week:
- 1 line on what happened this week (deployments, projects, any red days / no-JIRA).

This Month:
- 1 line on the month so far.

Overview:
- 1-2 lines on overall deployment health for the period.

Trend:
- 1-2 lines: is health improving or worsening month-over-month (deployments, JIRA compliance, red days)?

Needs Attention:
- The projects flagged "dirty" or "attention", each with the specific reason (no-JIRA patches, churn, failures, low JIRA compliance). If none, write "- None.".

Well Maintained:
- The projects flagged "good", briefly why. If none, write "- None.".

Recommendations:
- 2-4 concrete, actionable suggestions (e.g. enforce JIRA IDs on a project, investigate churn on a specific date).`

// GetAnalyticsNarrative asks Parson to analyze the computed insights for a period.
func GetAnalyticsNarrative(c fiber.Ctx) error {
	var settings database.Settings
	database.DB.First(&settings)
	if !settings.AIEnabled {
		return c.Status(400).JSON(fiber.Map{"error": "AI is disabled. Enable Parson in Settings."})
	}

	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))
	periodDeps, err := deploymentsForPeriod(month, year)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load deployments"})
	}
	recentDeps, err := deploymentsSince(time.Now().AddDate(0, -6, 0))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load recent deployments"})
	}
	insights := buildInsights(periodDeps, month, year)
	addTimeframe(&insights, recentDeps)
	if insights.TotalDeploys == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "No deployments in this period to analyze."})
	}

	prompt := buildAnalyticsPrompt(insights)

	client := aiClientFromSettings()
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()

	systemPrompt := analyticsSystemPrompt
	temperature := defaultTemperature
	if settings.ParsonTemperature > 0 {
		temperature = settings.ParsonTemperature
	}

	started := time.Now()
	out, err := client.Generate(ctx, prompt, ai.GenerateOptions{System: systemPrompt, Temperature: temperature})
	if err != nil {
		return c.Status(503).JSON(fiber.Map{"error": "AI analysis failed: " + err.Error()})
	}
	out = stripThinking(out)

	return c.JSON(fiber.Map{
		"narrative":          out,
		"generation_seconds": time.Since(started).Seconds(),
		"model":              client.Model,
		"insights":           insights,
	})
}

// buildAnalyticsPrompt renders the computed insights into Parson's input.
func buildAnalyticsPrompt(in InsightsResponse) string {
	var b strings.Builder
	fmt.Fprintf(&b, "PERIOD: %s\n", in.Period)
	fmt.Fprintf(&b, "THIS WEEK: %d deployments, %d projects, %d no-JIRA, %d red days (2+/day)\n",
		in.ThisWeek.Deployments, in.ThisWeek.Projects, in.ThisWeek.DirtyPatches, in.ThisWeek.RedDays)
	fmt.Fprintf(&b, "THIS MONTH: %d deployments, %d projects, %d no-JIRA, %d red days\n",
		in.ThisMonth.Deployments, in.ThisMonth.Projects, in.ThisMonth.DirtyPatches, in.ThisMonth.RedDays)
	fmt.Fprintf(&b, "TOTALS (%s): %d deployments across %d projects, %d%% success, %d%% JIRA compliance, %d no-JIRA patches\n\n",
		in.Period, in.TotalDeploys, in.TotalProjects, in.SuccessRate, in.JiraCompliance, in.DirtyPatches)
	b.WriteString("PER-PROJECT HEALTH:\n")
	for _, p := range in.Projects {
		fmt.Fprintf(&b, "- %s [%s, score %d/100]: %d deploys, %d%% success, %d%% JIRA, %d no-JIRA, %d churn-days, %d failed",
			p.Name, p.Classification, p.HealthScore, p.Deployments, p.SuccessRate, p.JiraCompliance, p.DirtyPatches, p.HighChurnDays, p.Failed)
		if p.BusiestDayCount >= 2 {
			fmt.Fprintf(&b, " (busiest: %d on %s)", p.BusiestDayCount, p.BusiestDay)
		}
		b.WriteString("\n")
	}

	b.WriteString("\nBY DEVELOPER:\n")
	for _, d := range in.Developers {
		fmt.Fprintf(&b, "- %s: %d deploys across %d project(s), %d%% success, %d%% JIRA, %d failed\n",
			d.Name, d.Deployments, d.Projects, d.SuccessRate, d.JiraCompliance, d.Failed)
	}

	b.WriteString("\nBY ENVIRONMENT:\n")
	for _, e := range in.Environments {
		fmt.Fprintf(&b, "- %s: %d deploys, %d%% success, %d failed, %d no-JIRA\n",
			e.Name, e.Deployments, e.SuccessRate, e.Failed, e.NoJira)
	}

	b.WriteString("\nMONTHLY TREND (last 6 months):\n")
	for _, m := range in.MonthlyTrends {
		fmt.Fprintf(&b, "- %s: %d deploys, %d%% success, %d%% JIRA, %d red days\n",
			m.Label, m.Deployments, m.SuccessRate, m.JiraCompliance, m.RedDays)
	}
	return b.String()
}

func periodLabel(month, year int) string {
	if month == 0 && year == 0 {
		return "All time"
	}
	if month == 0 {
		return strconv.Itoa(year)
	}
	t := time.Date(2000, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	if year == 0 {
		return t.Format("January")
	}
	return t.Format("January") + " " + strconv.Itoa(year)
}
