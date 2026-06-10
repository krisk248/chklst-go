package handlers

import (
	"chklst-go/internal/auth"

	"github.com/gofiber/fiber/v3"
)

// AuthStatus tells the frontend whether auth is on and whether the caller is logged in.
func AuthStatus(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"auth_enabled":  auth.Enabled(),
		"authenticated": auth.IsAuthenticated(c),
	})
}

// Login verifies the password and sets a session cookie.
func Login(c fiber.Ctx) error {
	var req struct {
		Password string `json:"password"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}
	if !auth.AllowLoginAttempt(c.IP()) {
		return c.Status(429).JSON(fiber.Map{"error": "Too many attempts — wait a minute and try again."})
	}
	if !auth.CheckPassword(req.Password) {
		return c.Status(401).JSON(fiber.Map{"error": "Incorrect password"})
	}
	auth.SetSessionCookie(c)
	return c.JSON(fiber.Map{"ok": true})
}

// Logout clears the session cookie.
func Logout(c fiber.Ctx) error {
	auth.ClearSessionCookie(c)
	return c.JSON(fiber.Map{"ok": true})
}
