package handlers

import (
	"fmt"
	"errors"
	"strings"
	"time"

	"chklst-go/internal/database"
	"chklst-go/internal/mail"

	"github.com/gofiber/fiber/v3"
)

// smtpConfig builds a mail.Config from the stored settings. Optional toOverride /
// ccOverride (comma-separated) replace the configured recipients when non-empty.
func smtpConfig(s database.Settings, toOverride, ccOverride string) mail.Config {
	to := s.SMTPTo
	if strings.TrimSpace(toOverride) != "" {
		to = toOverride
	}
	cc := s.SMTPCc
	if strings.TrimSpace(ccOverride) != "" {
		cc = ccOverride
	}
	return mail.Config{
		Host:     s.SMTPHost,
		Port:     s.SMTPPort,
		User:     s.SMTPUser,
		Password: s.SMTPPassword,
		From:     s.SMTPFrom,
		To:       mail.SplitAddrs(to),
		Cc:       mail.SplitAddrs(cc),
		Security: s.SMTPSecurity,
	}
}

// TestEmail sends a small test message to verify the SMTP configuration. It goes to
// SMTPTestTo (the dedicated test recipient) if set, otherwise the configured To, and
// never includes the Cc list — so test runs don't reach the real recipients.
func TestEmail(c fiber.Ctx) error {
	var s database.Settings
	database.DB.First(&s)

	testTo := s.SMTPTestTo
	if strings.TrimSpace(testTo) == "" {
		testTo = s.SMTPTo
	}

	cfg := smtpConfig(s, testTo, "")
	cfg.Cc = nil // a test run must not reach the real Cc list
	cfg.Headers = map[string]string{"X-Parson-AI": "PAi", "X-Mailer": "chklst-Parson"}
	if err := cfg.Send("chklst SMTP test (PAi)", "This is a test email from chklst / Parson. If you received this, SMTP is configured correctly.\n\n· PAi"); err != nil {
		return c.Status(400).JSON(fiber.Map{"ok": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"ok": true, "to": cfg.To})
}

// parsonFooter is the small visible "PAi" tag appended to AI-written mail.
const parsonFooter = "\n\n· PAi"

// sendSummaryRow emails a generated report and records the outcome on the row.
// Shared by the HTTP handler and the scheduler. Adds the PAi silent tag (header)
// plus a small visible "· PAi" marker so AI-written mail is easy to filter.
func sendSummaryRow(summary *database.DailySummary, s database.Settings) error {
	if strings.TrimSpace(summary.GeneratedBody) == "" {
		return errors.New("nothing to send — generate the report first")
	}

	cfg := smtpConfig(s, summary.Recipients, "")
	cfg.Headers = map[string]string{
		"X-Parson-AI": "PAi",        // silent, filterable tag
		"X-Mailer":    "chklst-Parson",
	}
	subject := summary.GeneratedSubject
	if subject == "" {
		subject = "Daily Activity Report – " + niceDate(summary.Date)
	}

	// Retry transient failures (DNS blips, connection resets) before giving up:
	// 3 attempts with short backoff. The scheduler adds its own 5-min retries on top.
	var err error
	for attempt, wait := 1, 2*time.Second; attempt <= 3; attempt, wait = attempt+1, wait*3 {
		err = cfg.Send(subject, summary.GeneratedBody+parsonFooter)
		if err == nil {
			break
		}
		if attempt < 3 {
			time.Sleep(wait)
		}
	}
	if err != nil {
		summary.Status = "failed"
		summary.SendError = err.Error()
		if serr := database.DB.Save(summary).Error; serr != nil {
			return errors.Join(err, serr)
		}
		return err
	}

	now := time.Now()
	summary.SentAt = &now
	summary.Status = "sent"
	summary.SendError = ""
	if err := database.DB.Save(summary).Error; err != nil {
		// The email went out, but recording it failed — surface so the caller
		// (and the scheduler) don't think send fully succeeded silently.
		return fmt.Errorf("email sent but failed to record status: %w", err)
	}
	return nil
}

// SendSummary emails a generated daily report and records the result on the row.
func SendSummary(c fiber.Ctx) error {
	id := c.Params("id")
	var summary database.DailySummary
	if err := database.DB.First(&summary, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Summary not found"})
	}

	var s database.Settings
	database.DB.First(&s)

	if err := sendSummaryRow(&summary, s); err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "Send failed: " + err.Error()})
	}
	return c.JSON(fiber.Map{"ok": true, "summary": summary})
}
