package handlers

import (
	"context"
	"strings"
	"time"

	"chklst-go/internal/ai"
	"chklst-go/internal/database"

	"github.com/gofiber/fiber/v3"
)

// aiClientFromSettings builds an Ollama client from the stored settings,
// applying defaults for any unset fields.
func aiClientFromSettings() *ai.Client {
	var s database.Settings
	database.DB.First(&s)
	url := s.OllamaURL
	if url == "" {
		url = defaultOllamaURL()
	}
	model := s.AIModel
	if model == "" {
		model = ai.DefaultModel
	}
	// CPU inference of a 12B model is slow; give it room.
	return ai.NewClient(url, model, 5*time.Minute)
}

// modelGuidance maps a known model tag to a short recommendation, so the Settings
// UI can explain "what to use" for this CPU laptop. See docs/AI_MODEL_GUIDE.md.
var modelGuidance = map[string]string{
	"gemma4:e4b": "Recommended — lightweight 4B, fastest on CPU, keeps patch detail.",
	"gemma4:e2b": "Smallest/fastest, but weaker writing — good for quick tests.",
	"gemma4:12b": "Dense 12B — stronger & concise; viable now that thinking is disabled.",
	"gemma4:26b": "MoE (~4B active) — near-frontier quality at ~4B speed; needs 18GB RAM.",
	"gemma4:31b": "Dense 31B — best quality but slow on CPU.",
	"gemma3:12b": "Previous-gen 12B — works, but Gemma 4 is preferred.",
}

func guidanceFor(name string) string {
	for tag, note := range modelGuidance {
		if name == tag || strings.HasPrefix(name, tag) {
			return note
		}
	}
	return ""
}

// suggestedModels are good options to pull even if not installed yet — surfaced in
// the UI so the user knows "what all I can use".
var suggestedModels = []fiber.Map{
	{"name": "gemma4:e4b", "size_hint": "9.6 GB", "note": modelGuidance["gemma4:e4b"]},
	{"name": "gemma4:12b", "size_hint": "7.6 GB", "note": modelGuidance["gemma4:12b"]},
	{"name": "gemma4:26b", "size_hint": "18 GB", "note": modelGuidance["gemma4:26b"]},
	{"name": "gemma4:e2b", "size_hint": "7.2 GB", "note": modelGuidance["gemma4:e2b"]},
}

// ListAIModels returns the models installed locally (with guidance) plus a curated
// list of suggestions the user could pull.
func ListAIModels(c fiber.Ctx) error {
	client := aiClientFromSettings()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	models, err := client.ListModels(ctx)
	if err != nil {
		return c.Status(503).JSON(fiber.Map{
			"error":      "Could not reach Ollama at " + client.BaseURL,
			"installed":  []any{},
			"suggested":  suggestedModels,
		})
	}

	installed := make([]fiber.Map, 0, len(models))
	for _, m := range models {
		installed = append(installed, fiber.Map{
			"name":    m.Name,
			"size":    m.Size,
			"size_gb": float64(m.Size) / 1e9,
			"note":    guidanceFor(m.Name),
		})
	}
	return c.JSON(fiber.Map{
		"installed": installed,
		"suggested": suggestedModels,
		"base_url":  client.BaseURL,
	})
}

// GetParsonDefaults returns the built-in agent defaults so the Settings UI can show
// or restore them (the "Reset to default" button for the system prompt).
func GetParsonDefaults(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"system_prompt": summarySystemPrompt,
		"temperature":   defaultTemperature,
	})
}

// TestAI verifies end-to-end connectivity: server reachable, model present, and a
// tiny generation succeeds. Used by the Settings page "Test AI" button.
func TestAI(c fiber.Ctx) error {
	client := aiClientFromSettings()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	version, err := client.Ping(ctx)
	if err != nil {
		return c.Status(503).JSON(fiber.Map{
			"ok":    false,
			"stage": "connect",
			"error": err.Error(),
		})
	}

	hasModel, err := client.HasModel(ctx)
	if err != nil {
		return c.Status(503).JSON(fiber.Map{
			"ok":      false,
			"stage":   "list_models",
			"version": version,
			"error":   err.Error(),
		})
	}
	if !hasModel {
		return c.Status(404).JSON(fiber.Map{
			"ok":      false,
			"stage":   "model_missing",
			"version": version,
			"model":   client.Model,
			"error":   "Model not pulled yet. Run: ollama pull " + client.Model,
		})
	}

	// Minimal generation to confirm the model actually responds.
	out, err := client.Generate(ctx, "Reply with exactly the word: OK", ai.GenerateOptions{
		Temperature: 0,
	})
	if err != nil {
		return c.Status(503).JSON(fiber.Map{
			"ok":      false,
			"stage":   "generate",
			"version": version,
			"model":   client.Model,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"ok":       true,
		"version":  version,
		"model":    client.Model,
		"base_url": client.BaseURL,
		"sample":   out,
	})
}
