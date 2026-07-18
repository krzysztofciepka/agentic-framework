package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/krzysztofciepka/agentic-framework/internal/model"
	"github.com/krzysztofciepka/agentic-framework/internal/tool"
)

type ParamProperty struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type ToolSchema struct {
	Type       string                   `json:"type"`
	Properties map[string]ParamProperty `json:"properties"`
	Required   []string                 `json:"required"`
}

type Message struct {
	Role       string      `json:"role"`
	Content    string      `json:"content,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
	ToolCalls  []*ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	ID       string        `json:"id"`
	Type     string        `json:"type"`
	Index    int           `json:"index"`
	Function *FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolDef struct {
	Type     string       `json:"type"`
	Function *FunctionDef `json:"function"`
}

type FunctionDef struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  ToolSchema `json:"parameters"`
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Tools       []ToolDef `json:"tools,omitempty"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
	Stream      bool      `json:"stream"`
}

type Choice struct {
	Index   int     `json:"index"`
	Message Message `json:"message"`
	Delta   Delta   `json:"delta"`
}

type Delta struct {
	Role      string      `json:"role,omitempty"`
	Content   string      `json:"content,omitempty"`
	ToolCalls []*ToolCall `json:"tool_calls,omitempty"`
}

type ChatResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
}

type StreamChunk struct {
	Choices []Choice `json:"choices"`
}

type Client struct {
	baseURL string
	apiKey  string
	model   string
	http    *http.Client
}

func NewClient(baseURL, apiKey, model string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		model:   model,
		http: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

func (c *Client) Chat(ctx context.Context, messages []Message, tools []tool.Tool, temperature float64, maxTokens int) ([]Choice, error) {
	var toolDefs []ToolDef
	for _, t := range tools {
		td := ToolDef{
			Type: "function",
			Function: &FunctionDef{
				Name:        t.Name(),
				Description: t.Description(),
			},
		}
		params := t.Parameters()
		td.Function.Parameters = ToolSchema{
			Type:       params.Type,
			Properties: make(map[string]ParamProperty, len(params.Properties)),
			Required:   params.Required,
		}
		for k, v := range params.Properties {
			td.Function.Parameters.Properties[k] = ParamProperty{
				Type:        v.Type,
				Description: v.Description,
			}
		}
		toolDefs = append(toolDefs, td)
	}

	req := ChatRequest{
		Model:       c.model,
		Messages:    messages,
		Tools:       toolDefs,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Stream:      false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return chatResp.Choices, nil
}

func (c *Client) ChatStream(ctx context.Context, messages []Message, tools []tool.Tool, temperature float64, maxTokens int, ch chan<- StreamChunk) error {
	defer close(ch)

	var toolDefs []ToolDef
	for _, t := range tools {
		td := ToolDef{
			Type: "function",
			Function: &FunctionDef{
				Name:        t.Name(),
				Description: t.Description(),
			},
		}
		params := t.Parameters()
		td.Function.Parameters = ToolSchema{
			Type:       params.Type,
			Properties: make(map[string]ParamProperty, len(params.Properties)),
			Required:   params.Required,
		}
		for k, v := range params.Properties {
			td.Function.Parameters.Properties[k] = ParamProperty{
				Type:        v.Type,
				Description: v.Description,
			}
		}
		toolDefs = append(toolDefs, td)
	}

	req := ChatRequest{
		Model:       c.model,
		Messages:    messages,
		Tools:       toolDefs,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Stream:      true,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			return nil
		}

		var chunk StreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- chunk:
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan response: %w", err)
	}

	return nil
}

func ModelMessagesToLLM(msgs []model.Message) []Message {
	result := make([]Message, len(msgs))
	for i, m := range msgs {
		result[i] = Message{
			Role:       m.Role,
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		}
	}
	return result
}
