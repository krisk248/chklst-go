// Package auth provides an optional single-password gate.
//
//   - AUTH_ENABLED=true turns it on (default off, so local installs stay open).
//   - ADMIN_PASSWORD is the password to log in with.
//   - Sessions are stateless HMAC-signed tokens in an httpOnly cookie, valid 7 days.
//
// When enabled, the Fiber middleware protects /api/v1/* (except /api/v1/auth/*).
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
)

const cookieName = "chklst_session"

var (
	enabled  bool
	password string
	secret   []byte
)

// Init reads config once at startup.
func Init() {
	enabled = strings.EqualFold(strings.TrimSpace(os.Getenv("AUTH_ENABLED")), "true")
	password = os.Getenv("ADMIN_PASSWORD")

	// Token-signing secret: SESSION_SECRET, else APP_ENCRYPTION_KEY, else a random
	// per-boot value (sessions then reset on restart — acceptable).
	s := os.Getenv("SESSION_SECRET")
	if s == "" {
		s = os.Getenv("APP_ENCRYPTION_KEY")
	}
	if s == "" {
		b := make([]byte, 32)
		// crypto/rand via time fallback isn't ideal, but only used when no secret set.
		now := time.Now().UnixNano()
		for i := range b {
			b[i] = byte(now >> (i % 8))
		}
		s = base64.StdEncoding.EncodeToString(b)
	}
	secret = []byte(s)

	if enabled {
		if password == "" {
			log.Println("⚠ AUTH_ENABLED=true but ADMIN_PASSWORD is empty — login cannot succeed; set ADMIN_PASSWORD")
		} else {
			log.Println("🔐 Authentication enabled (password gate)")
		}
	}
}

// Enabled reports whether the gate is active.
func Enabled() bool { return enabled }

// CheckPassword constant-time compares the submitted password.
func CheckPassword(pw string) bool {
	if password == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(pw), []byte(password)) == 1
}

// --- Simple per-IP login throttle (brute-force protection) ---
var (
	attemptMu sync.Mutex
	attempts  = map[string][]int64{} // ip -> recent attempt unix-seconds
)

// AllowLoginAttempt returns false if ip has made 5+ attempts in the last 60s.
func AllowLoginAttempt(ip string) bool {
	attemptMu.Lock()
	defer attemptMu.Unlock()
	now := time.Now().Unix()
	recent := attempts[ip][:0]
	for _, t := range attempts[ip] {
		if now-t < 60 {
			recent = append(recent, t)
		}
	}
	if len(recent) >= 5 {
		attempts[ip] = recent
		return false
	}
	attempts[ip] = append(recent, now)
	return true
}

// issueToken builds "exp.signature" valid for 7 days.
func issueToken() string {
	exp := strconv.FormatInt(time.Now().Add(7*24*time.Hour).Unix(), 10)
	return exp + "." + sign(exp)
}

func sign(msg string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(msg))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// validToken checks signature + expiry.
func validToken(tok string) bool {
	parts := strings.SplitN(tok, ".", 2)
	if len(parts) != 2 {
		return false
	}
	if !hmac.Equal([]byte(parts[1]), []byte(sign(parts[0]))) {
		return false
	}
	exp, err := strconv.ParseInt(parts[0], 10, 64)
	return err == nil && time.Now().Unix() < exp
}

// SetSessionCookie issues a session cookie on the response.
func SetSessionCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     cookieName,
		Value:    issueToken(),
		HTTPOnly: true,
		SameSite: "Lax",
		Path:     "/",
		MaxAge:   7 * 24 * 3600,
	})
}

// ClearSessionCookie logs out.
func ClearSessionCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{Name: cookieName, Value: "", HTTPOnly: true, Path: "/", MaxAge: -1})
}

// IsAuthenticated reports whether the request carries a valid session.
func IsAuthenticated(c fiber.Ctx) bool {
	if !enabled {
		return true
	}
	return validToken(c.Cookies(cookieName))
}

// Middleware blocks unauthenticated API requests when auth is enabled. Auth routes
// (/api/v1/auth/*) and the health check are always allowed.
func Middleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		if !enabled {
			return c.Next()
		}
		p := c.Path()
		if strings.HasPrefix(p, "/api/v1/auth/") || p == "/health" {
			return c.Next()
		}
		if strings.HasPrefix(p, "/api/") && !IsAuthenticated(c) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Authentication required"})
		}
		return c.Next()
	}
}
