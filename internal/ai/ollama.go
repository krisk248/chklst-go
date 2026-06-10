// Package ai provides a thin client for a local Ollama server, used to generate
// the daily activity summary and analytics narratives with a local LLM (Gemma 4).
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultBaseURL is where a local Ollama listens. Inside Docker this is overridden
// (via Settings / OLLAMA_URL) with http://host.docker.internal:11434.
const DefaultBaseURL = "http://localhost:11434"

// DefaultModel is the model we standardize on. On this CPU-only laptop, benchmarking
// showed gemma4:e4b (a 4B "edge" variant) generates the daily summary ~4.5x faster
// than the dense gemma4:12b while keeping more patch-level detail — the best
// speed/quality trade here. See docs/AI_MODEL_GUIDE.md. Override per-install in Settings.
const DefaultModel = "gemma4:e4b"

// Client talks to an Ollama server.
type Client struct {
	BaseURL string
	Model   string
	http    *http.Client
}

// NewClient builds a client. Empty baseURL/model fall back to the defaults.
// timeout bounds a single generation; CPU inference of a 12B model is slow, so
// callers should pass a generous value (minutes), not the usual few seconds.
func NewClient(baseURL, model string, timeout time.Duration) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	if model == "" {
		model = DefaultModel
	}
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Model:   model,
		http:    &http.Client{Timeout: timeout},
	}
}

// Unload asks Ollama to evict the model from memory immediately (keep_alive: 0),
// freeing RAM right after a scheduled generation instead of waiting for the idle
// timeout. Best-effort: errors are returned but callers typically just log them.
func (c *Client) Unload(ctx context.Context) error {
	body, _ := json.Marshal(map[string]any{"model": c.Model, "keep_alive": 0})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// versionResponse is the shape of GET /api/version.
type versionResponse struct {
	Version string `json:"version"`
}

// Ping verifies the server is reachable and returns its version.
func (c *Client) Ping(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/api/version", nil)
	if err != nil {
		return "", err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama unreachable at %s: %w", c.BaseURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}
	var v versionResponse
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return "", err
	}
	return v.Version, nil
}

// ModelInfo describes a locally-available model.
type ModelInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// tagsResponse is the shape of GET /api/tags.
type tagsResponse struct {
	Models []ModelInfo `json:"models"`
}

// ListModels returns the models pulled into the local Ollama.
func (c *Client) ListModels(ctx context.Context) ([]ModelInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/api/tags", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama unreachable at %s: %w", c.BaseURL, err)
	}
	defer resp.Body.Close()
	var t tagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, err
	}
	return t.Models, nil
}

// HasModel reports whether the configured model is present locally.
func (c *Client) HasModel(ctx context.Context) (bool, error) {
	models, err := c.ListModels(ctx)
	if err != nil {
		return false, err
	}
	for _, m := range models {
		// tags may be "gemma4:12b" or "gemma4:12b-..." — match on the base.
		if m.Name == c.Model || strings.HasPrefix(m.Name, c.Model) {
			return true, nil
		}
	}
	return false, nil
}

// generateRequest is the body for POST /api/generate (non-streaming).
type generateRequest struct {
	Model   string         `json:"model"`
	Prompt  string         `json:"prompt"`
	System  string         `json:"system,omitempty"`
	Stream  bool           `json:"stream"`
	Think   *bool          `json:"think,omitempty"` // false disables "thinking" models (much faster)
	Options map[string]any `json:"options,omitempty"`
}

type generateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

// GenerateOptions tunes a single generation.
type GenerateOptions struct {
	System      string
	Temperature float64 // 0 = deterministic-ish; good for structured summaries
	// Think controls "thinking" models (e.g. gemma4:12b). For structured formatting
	// tasks we disable it: it otherwise burns hundreds of reasoning tokens (minutes
	// on CPU) for no benefit. Defaults to disabled.
	Think bool
}

// Generate runs a single non-streaming completion and returns the model's text.
// It returns a friendly error when the server is down or the model is missing,
// so callers can surface "AI unavailable" without leaking transport details.
func (c *Client) Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	if strings.TrimSpace(prompt) == "" {
		return "", errors.New("empty prompt")
	}

	think := opts.Think
	body := generateRequest{
		Model:  c.Model,
		Prompt: prompt,
		System: opts.System,
		Stream: false,
		Think:  &think, // explicitly false by default to skip slow reasoning
		Options: map[string]any{
			"temperature": opts.Temperature,
		},
	}

	out, err := c.doGenerate(ctx, body)
	if err != nil {
		// Some non-thinking models reject the `think` field; retry once without it.
		if strings.Contains(strings.ToLower(err.Error()), "think") {
			body.Think = nil
			return c.doGenerate(ctx, body)
		}
		return "", err
	}
	return out, nil
}

// doGenerate performs a single /api/generate call.
func (c *Client) doGenerate(ctx context.Context, body generateRequest) (string, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/generate", bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama generation failed (is `ollama serve` running at %s?): %w", c.BaseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}

	var gr generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return "", err
	}
	if gr.Error != "" {
		return "", fmt.Errorf("ollama error: %s", gr.Error)
	}
	return strings.TrimSpace(gr.Response), nil
}
