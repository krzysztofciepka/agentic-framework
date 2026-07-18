package tool

import (
	"context"
	"errors"
	"fmt"
	"sync"
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

type Tool interface {
	Name() string
	Description() string
	Category() string
	Parameters() ToolSchema
	Execute(ctx context.Context, args map[string]any) (string, error)
}

type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

func (r *Registry) Register(t Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[t.Name()] = t
}

func (r *Registry) Get(name string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("tool %q not found", name)
	}
	return t, nil
}

func (r *Registry) All() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tools := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, t)
	}
	return tools
}

func DefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register(&WebSearch{})
	r.Register(&WebFetch{})
	r.Register(&RunCmd{})
	r.Register(&FileRead{})
	r.Register(&FileWrite{})
	r.Register(&Execute{})
	return r
}

var ErrToolNotFound = errors.New("tool not found")
