package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type WebSearch struct{}

func (w *WebSearch) Name() string        { return "web_search" }
func (w *WebSearch) Description() string { return "Search the web using Brave Search API" }
func (w *WebSearch) Category() string    { return "web" }

func (w *WebSearch) Parameters() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]ParamProperty{
			"query": {Type: "string", Description: "The search query"},
			"count": {Type: "integer", Description: "Number of results (default 10, max 20)"},
		},
		Required: []string{"query"},
	}
}

type braveWebResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

type braveWeb struct {
	Results []braveWebResult `json:"results"`
}

type braveResponse struct {
	Web braveWeb `json:"web"`
}

func (w *WebSearch) Execute(ctx context.Context, args map[string]any) (string, error) {
	apiKey := os.Getenv("BRAVE_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("BRAVE_API_KEY environment variable is not set")
	}

	query, ok := args["query"].(string)
	if !ok || query == "" {
		return "", fmt.Errorf("query parameter is required and must be a non-empty string")
	}

	count := 10
	if c, ok := args["count"]; ok {
		switch v := c.(type) {
		case float64:
			count = int(v)
		case int:
			count = v
		case string:
			parsed, err := strconv.Atoi(v)
			if err != nil {
				return "", fmt.Errorf("count must be an integer, got %q", v)
			}
			count = parsed
		}
	}
	if count < 1 {
		count = 1
	}
	if count > 20 {
		count = 20
	}

	reqURL := fmt.Sprintf("https://api.search.brave.com/res/v1/web/search?q=%s&count=%d",
		url.QueryEscape(query), count)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("brave search API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result braveResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(result.Web.Results) == 0 {
		return "No results found.", nil
	}

	var output string
	for i, r := range result.Web.Results {
		output += fmt.Sprintf("%d. %s\n   %s\n   %s\n\n", i+1, r.Title, r.URL, r.Description)
	}
	return output, nil
}
