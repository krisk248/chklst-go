package middleware

import "github.com/gofiber/fiber/v3"

// SecurityHeaders sets conservative security headers on every response.
func SecurityHeaders() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "SAMEORIGIN")
		c.Set("Referrer-Policy", "no-referrer")
		c.Set("X-XSS-Protection", "0") // modern browsers; rely on CSP/escaping
		return c.Next()
	}
}
