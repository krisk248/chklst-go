package handlers

import (
	"time"
	"strings"

	"chklst-go/internal/database"

	"github.com/gofiber/fiber/v3"
)

// ListHolidays returns all configured holidays (soonest first).
func ListHolidays(c fiber.Ctx) error {
	var holidays []database.Holiday
	database.DB.Order("date asc").Find(&holidays)
	return c.JSON(holidays)
}

// AddHoliday adds (or updates the name of) a holiday date.
func AddHoliday(c fiber.Ctx) error {
	var req database.Holiday
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	req.Date = strings.TrimSpace(req.Date)
	if req.Date == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Date is required (YYYY-MM-DD)"})
	}
	if _, err := time.ParseInLocation("2006-01-02", req.Date, time.Local); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid date — use YYYY-MM-DD"})
	}

	var existing database.Holiday
	if database.DB.Where("date = ?", req.Date).First(&existing).Error == nil {
		existing.Name = req.Name
		if err := database.DB.Save(&existing).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to update holiday"})
		}
		return c.JSON(existing)
	}
	if err := database.DB.Create(&req).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to add holiday"})
	}
	return c.Status(201).JSON(req)
}

// DeleteHoliday removes a holiday by id.
func DeleteHoliday(c fiber.Ctx) error {
	id := c.Params("id")
	if err := database.DB.Delete(&database.Holiday{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete holiday"})
	}
	return c.SendStatus(204)
}

// isHoliday reports whether dateStr ("2006-01-02") is a configured holiday,
// returning the holiday name when so.
func isHoliday(dateStr string) (bool, string) {
	var h database.Holiday
	if database.DB.Where("date = ?", dateStr).First(&h).Error == nil {
		return true, h.Name
	}
	return false, ""
}
