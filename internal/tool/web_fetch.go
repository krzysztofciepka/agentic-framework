package tool

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type WebFetch struct{}

func (w *WebFetch) Name() string        { return "web_fetch" }
func (w *WebFetch) Description() string { return "Fetch content from a URL (text/markdown/html)" }
func (w *WebFetch) Category() string    { return "web" }

func (w *WebFetch) Parameters() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]ParamProperty{
			"url":    {Type: "string", Description: "The URL to fetch"},
			"format": {Type: "string", Description: "Response format: text, markdown, or html (default: text)"},
		},
		Required: []string{"url"},
	}
}

func (w *WebFetch) Execute(ctx context.Context, args map[string]any) (string, error) {
	rawURL, ok := args["url"].(string)
	if !ok || rawURL == "" {
		return "", fmt.Errorf("url parameter is required and must be a non-empty string")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "agentic-framework/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch request: %w", err)
	}
	defer resp.Body.Close()

	const maxBody = 1 << 20
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	return strings.TrimSpace(string(body)), nil
}
