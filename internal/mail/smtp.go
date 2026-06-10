// Package mail sends the daily activity report over SMTP.
package mail

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// Config holds SMTP connection + addressing settings.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
	To       []string
	Cc       []string
	Security string            // "starttls" | "tls" | "none"
	Headers  map[string]string // extra headers, e.g. the X-Parson-AI tag
}

// Validate checks the minimum required fields.
func (c Config) Validate() error {
	if c.Host == "" {
		return errors.New("SMTP host is required")
	}
	if c.Port == 0 {
		return errors.New("SMTP port is required")
	}
	if c.From == "" {
		return errors.New("From address is required")
	}
	if len(c.To) == 0 {
		return errors.New("at least one recipient (To) is required")
	}
	return nil
}

// splitAddrs parses a comma/semicolon-separated address list, trimming blanks.
func SplitAddrs(s string) []string {
	fields := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ';' })
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if t := strings.TrimSpace(f); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// buildMessage assembles an RFC 5322 message.
func (c Config) buildMessage(subject, body string) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", c.From)
	fmt.Fprintf(&b, "To: %s\r\n", strings.Join(c.To, ", "))
	if len(c.Cc) > 0 {
		fmt.Fprintf(&b, "Cc: %s\r\n", strings.Join(c.Cc, ", "))
	}
	fmt.Fprintf(&b, "Subject: %s\r\n", subject)
	for k, v := range c.Headers {
		fmt.Fprintf(&b, "%s: %s\r\n", k, v)
	}
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	b.WriteString("\r\n")
	// Normalize line endings to CRLF for the body.
	b.WriteString(strings.ReplaceAll(strings.ReplaceAll(body, "\r\n", "\n"), "\n", "\r\n"))
	return []byte(b.String())
}

// Send delivers a plain-text email. It supports implicit TLS (port 465),
// STARTTLS (typically 587), or an unencrypted connection ("none").
func (c Config) Send(subject, body string) error {
	if err := c.Validate(); err != nil {
		return err
	}

	addr := net.JoinHostPort(c.Host, fmt.Sprintf("%d", c.Port))
	recipients := append(append([]string{}, c.To...), c.Cc...)
	msg := c.buildMessage(subject, body)

	var auth smtp.Auth
	if c.User != "" {
		auth = smtp.PlainAuth("", c.User, c.Password, c.Host)
	}

	security := strings.ToLower(strings.TrimSpace(c.Security))
	if security == "" {
		// Infer from port if not explicitly set.
		switch c.Port {
		case 465:
			security = "tls"
		case 25:
			security = "none"
		default:
			security = "starttls"
		}
	}

	if security == "tls" {
		return c.sendImplicitTLS(addr, auth, recipients, msg)
	}
	return c.sendPlainOrStartTLS(addr, auth, recipients, msg, security == "starttls")
}

// sendImplicitTLS dials a TLS connection directly (port 465 style).
func (c Config) sendImplicitTLS(addr string, auth smtp.Auth, recipients []string, msg []byte) error {
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 30 * time.Second}, "tcp", addr, &tls.Config{ServerName: c.Host})
	if err != nil {
		return fmt.Errorf("TLS dial failed: %w", err)
	}
	client, err := smtp.NewClient(conn, c.Host)
	if err != nil {
		return err
	}
	defer client.Close()
	return c.deliver(client, auth, recipients, msg)
}

// sendPlainOrStartTLS dials plainly and upgrades to STARTTLS when requested/available.
func (c Config) sendPlainOrStartTLS(addr string, auth smtp.Auth, recipients []string, msg []byte, useStartTLS bool) error {
	conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	client, err := smtp.NewClient(conn, c.Host)
	if err != nil {
		return err
	}
	defer client.Close()

	if useStartTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: c.Host}); err != nil {
				return fmt.Errorf("STARTTLS failed: %w", err)
			}
		} else {
			return errors.New("server does not support STARTTLS; set Security to 'none' or 'tls'")
		}
	}
	return c.deliver(client, auth, recipients, msg)
}

// deliver runs AUTH + the SMTP transaction on an established client.
func (c Config) deliver(client *smtp.Client, auth smtp.Auth, recipients []string, msg []byte) error {
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth failed: %w", err)
		}
	}
	if err := client.Mail(c.From); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}
	for _, r := range recipients {
		if err := client.Rcpt(r); err != nil {
			return fmt.Errorf("RCPT TO %s failed: %w", r, err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}
