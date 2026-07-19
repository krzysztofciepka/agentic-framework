package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/krzysztofciepka/agentic-framework/internal/llm"
	"github.com/krzysztofciepka/agentic-framework/internal/tool"
)

const maxToolLoops = 10

type Orchestrator struct {
	registry *tool.Registry
}

func NewOrchestrator(registry *tool.Registry) *Orchestrator {
	return &Orchestrator{registry: registry}
}

type StreamEvent struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Tool    string `json:"tool,omitempty"`
}

func (o *Orchestrator) Run(
	ctx context.Context,
	client *llm.Client,
	messages []llm.Message,
	toolNames []string,
	temperature float64,
	maxTokens int,
) ([]llm.Message, []llm.Message, error) {
	tools := o.resolveTools(toolNames)
	results := make([]llm.Message, 0)

	for range maxToolLoops {
		choices, err := client.Chat(ctx, messages, tools, temperature, maxTokens)
		if err != nil {
			return nil, nil, fmt.Errorf("chat: %w", err)
		}
		if len(choices) == 0 {
			return nil, nil, fmt.Errorf("no choices returned")
		}

		msg := choices[0].Message
		results = append(results, msg)
		messages = append(messages, msg)

		if len(msg.ToolCalls) == 0 {
			return results, messages, nil
		}

		for _, tc := range msg.ToolCalls {
			if tc.Function == nil {
				continue
			}
			toolMsg := o.executeTool(ctx, tc.Function.Name, tc.Function.Arguments, tc.ID)
			messages = append(messages, toolMsg)
		}
	}

	return nil, nil, fmt.Errorf("max tool loops (%d) exceeded", maxToolLoops)
}

func (o *Orchestrator) RunStream(
	ctx context.Context,
	client *llm.Client,
	messages []llm.Message,
	toolNames []string,
	temperature float64,
	maxTokens int,
	eventCh chan<- StreamEvent,
) ([]llm.Message, error) {
	defer close(eventCh)
	tools := o.resolveTools(toolNames)

	for range maxToolLoops {
		chunkCh := make(chan llm.StreamChunk)

		go func() {
			if err := client.ChatStream(ctx, messages, tools, temperature, maxTokens, chunkCh); err != nil {
				log.Printf("agent: stream error: %v", err)
			}
		}()

		var fullContent string
		toolCallAccum := make(map[int]*llm.ToolCall)

		for chunk := range chunkCh {
			for _, choice := range chunk.Choices {
				delta := choice.Delta

				if delta.Content != "" {
					fullContent += delta.Content
					eventCh <- StreamEvent{Type: "content", Content: delta.Content}
				}

				for _, tc := range delta.ToolCalls {
					if tc.Function == nil {
						continue
					}

					if existing, ok := toolCallAccum[tc.Index]; ok {
						if tc.Function.Name != "" {
							existing.Function.Name = tc.Function.Name
						}
						existing.Function.Arguments += tc.Function.Arguments
						if tc.ID != "" {
							existing.ID = tc.ID
						}
					} else {
						eventCh <- StreamEvent{Type: "tool_start", Tool: tc.Function.Name}
						toolCallAccum[tc.Index] = tc
					}
				}
			}
		}

		if len(toolCallAccum) == 0 {
			messages = append(messages, llm.Message{
				Role:    "assistant",
				Content: fullContent,
			})
			eventCh <- StreamEvent{Type: "done"}
			return messages, nil
		}

		toolCalls := make([]*llm.ToolCall, len(toolCallAccum))
		for idx := 0; idx < len(toolCalls); idx++ {
			toolCalls[idx] = toolCallAccum[idx]
		}

		assistantMsg := llm.Message{
			Role:      "assistant",
			Content:   fullContent,
			ToolCalls: toolCalls,
		}
		messages = append(messages, assistantMsg)

		for _, tc := range toolCalls {
			toolMsg := o.executeTool(ctx, tc.Function.Name, tc.Function.Arguments, tc.ID)
			messages = append(messages, toolMsg)
			toolContent, _ := toolMsg.Content.(string)
			eventCh <- StreamEvent{Type: "tool_end", Content: toolContent, Tool: tc.Function.Name}
		}
	}

	eventCh <- StreamEvent{Type: "error", Content: fmt.Sprintf("max tool loops (%d) exceeded", maxToolLoops)}
	return messages, nil
}

func (o *Orchestrator) executeTool(ctx context.Context, name, argsJSON, callID string) llm.Message {
	t, err := o.registry.Get(name)
	if err != nil {
		return llm.Message{
			Role:       "tool",
			ToolCallID: callID,
			Content:    fmt.Sprintf(`{"error": "tool %q not found"}`, name),
		}
	}

	var args map[string]any
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return llm.Message{
			Role:       "tool",
			ToolCallID: callID,
			Content:    fmt.Sprintf(`{"error": "invalid arguments: %s"}`, err.Error()),
		}
	}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return llm.Message{
			Role:       "tool",
			ToolCallID: callID,
			Content:    fmt.Sprintf(`{"error": %q}`, err.Error()),
		}
	}

	content := result
	if content == "" {
		content = `{"result": "ok"}`
	}

	return llm.Message{
		Role:       "tool",
		ToolCallID: callID,
		Content:    content,
	}
}

func (o *Orchestrator) resolveTools(names []string) []tool.Tool {
	resolved := make([]tool.Tool, 0, len(names))
	for _, name := range names {
		t, err := o.registry.Get(name)
		if err != nil {
			log.Printf("agent: tool %q not found in registry, skipping", name)
			continue
		}
		resolved = append(resolved, t)
	}
	return resolved
}
