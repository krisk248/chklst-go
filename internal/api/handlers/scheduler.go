package handlers

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"chklst-go/internal/database"
)

// sendAttempts tracks in-memory retry pacing for auto-send, keyed by date.
// (Reset on restart, which is fine: the send conditions are re-derived from DB state.)
var (
	sendAttemptMu   sync.Mutex
	lastSendAttempt = map[string]time.Time{}
)

// StartScheduler launches Parson's daily-report scheduler in a background goroutine.
//
// Design (replaces the old exact-minute matching):
//   - GENERATE: fires any tick where now >= generate-time and today's report hasn't
//     been generated since that time. This survives app restarts (catch-up), skipped
//     ticks (sleep/suspend), and never clobbers a report the user already reviewed.
//   - SEND: fires any tick where now >= send-time and the report is generated but
//     unsent. Failures are retried every 5 minutes until 23:00 (transient SMTP/DNS
//     blips no longer kill the daily email).
//   - All actions are idempotent off DailySummary status fields, and generation is
//     serialized by genMu (shared with the manual Generate endpoint).
func StartScheduler() {
	// DISABLE_SCHEDULER=true turns off auto-generate/auto-send entirely — used by the
	// dev instance (port 9000) so it never emails the team while we build against
	// a copy of real data.
	if strings.EqualFold(strings.TrimSpace(os.Getenv("DISABLE_SCHEDULER")), "true") {
		log.Println("🕒 Parson scheduler DISABLED (DISABLE_SCHEDULER=true)")
		return
	}
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		log.Println("🕒 Parson scheduler started")
		for {
			tick()
			<-ticker.C
		}
	}()
}

func tick() {
	var s database.Settings
	if err := database.DB.First(&s).Error; err != nil {
		return
	}
	if !s.SummaryScheduleEnabled {
		return
	}

	now := time.Now()
	today := now.Format("2006-01-02")

	genAt, gerr := atToday(s.SummaryGenerateTime, now)
	sendAt, serr := atToday(s.SummarySendTime, now)
	if gerr != nil || serr != nil {
		log.Printf("🕒 Parson: invalid schedule times (%q / %q) — fix in Settings", s.SummaryGenerateTime, s.SummarySendTime)
		return
	}
	if now.Before(genAt) && now.Before(sendAt) {
		return // nothing due yet today
	}

	// Skip inactive weekdays and holidays (log only when something would be due).
	if !weekdayActive(int(now.Weekday()), s.SummaryWeekdays) {
		return
	}
	if holiday, name := isHoliday(today); holiday {
		logOncePerDay("holiday-"+today, func() {
			log.Printf("🕒 Parson: %s is a holiday (%s) — skipping", today, name)
		})
		return
	}

	summary, err := getOrCreateSummary(today)
	if err != nil {
		return
	}

	// 1) GENERATE (incl. restart catch-up): due, AI on, not sent, not user-reviewed,
	// and not generated since the scheduled time.
	if s.AIEnabled && !now.Before(genAt) && summary.Status != "sent" && !summary.Reviewed &&
		(summary.GeneratedAt == nil || summary.GeneratedAt.Before(genAt)) {

		if !genMu.TryLock() {
			return // a generation is already running; next tick re-checks
		}
		late := now.Sub(genAt).Round(time.Minute)
		if late > 2*time.Minute {
			log.Printf("🕒 Parson: catch-up — generate was due %s ago (restart/sleep?)", late)
		}
		log.Printf("🕒 Parson: auto-generating report for %s (model %s)", today, s.AIModel)
		elapsed, model, gerr := generateSummaryRow(&summary, s, "", true)
		genMu.Unlock()
		if gerr != nil {
			log.Printf("🕒 Parson generate FAILED for %s: %v", today, gerr)
			return
		}
		log.Printf("🕒 Parson: generated %s in %.1fs with %s — locked", today, elapsed.Seconds(), model)

		// Free RAM immediately now that the report is done and locked.
		client := aiClientFromSettings()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if uerr := client.Unload(ctx); uerr != nil {
			log.Printf("🕒 Parson: model unload note: %v", uerr)
		} else {
			log.Printf("🕒 Parson: unloaded model to free memory")
		}
		cancel()
	}

	// 2) SEND (with retry until 23:00): due, auto-send on, generated, unsent.
	if s.SummaryAutoSend && !now.Before(sendAt) && now.Hour() < 23 &&
		summary.GeneratedBody != "" && summary.Status != "sent" {

		if !shouldAttemptSend(today, now) {
			return // pacing: wait before retrying a failed send
		}
		log.Printf("🕒 Parson: auto-sending report for %s (reviewed=%v) to %s",
			today, summary.Reviewed, recipientsLabel(summary, s))
		if serr := sendSummaryRow(&summary, s); serr != nil {
			log.Printf("🕒 Parson send FAILED for %s: %v — will retry in ~5 min (until 23:00)", today, serr)
		} else {
			log.Printf("🕒 Parson: sent %s ✓", today)
		}
	}
}

// shouldAttemptSend allows the first attempt immediately, then one retry per 5 minutes.
func shouldAttemptSend(date string, now time.Time) bool {
	sendAttemptMu.Lock()
	defer sendAttemptMu.Unlock()
	if last, ok := lastSendAttempt[date]; ok && now.Sub(last) < 5*time.Minute {
		return false
	}
	lastSendAttempt[date] = now
	return true
}

// atToday parses "HH:MM" onto today's date in local time.
func atToday(hhmm string, now time.Time) (time.Time, error) {
	t, err := time.ParseInLocation("15:04", strings.TrimSpace(hhmm), time.Local)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, time.Local), nil
}

// logOncePerDay suppresses repeated identical log lines (the scheduler ticks every minute).
var (
	logOnceMu   sync.Mutex
	logOnceSeen = map[string]bool{}
)

func logOncePerDay(key string, fn func()) {
	logOnceMu.Lock()
	defer logOnceMu.Unlock()
	if logOnceSeen[key] {
		return
	}
	logOnceSeen[key] = true
	fn()
}

// recipientsLabel describes where a report will go, for logging.
func recipientsLabel(summary database.DailySummary, s database.Settings) string {
	to := s.SMTPTo
	if strings.TrimSpace(summary.Recipients) != "" {
		to = summary.Recipients
	}
	return to
}

// weekdayActive reports whether weekday (0=Sun..6=Sat) is in the CSV list.
func weekdayActive(weekday int, csv string) bool {
	for _, p := range strings.Split(csv, ",") {
		if n, err := strconv.Atoi(strings.TrimSpace(p)); err == nil && n == weekday {
			return true
		}
	}
	return false
}
