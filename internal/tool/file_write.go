package tool

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type FileWrite struct{}

func (f *FileWrite) Name() string        { return "file_system_write" }
func (f *FileWrite) Description() string { return "Write content to a file (creates parent directories if needed)" }
func (f *FileWrite) Category() string    { return "file" }

func (f *FileWrite) Parameters() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]ParamProperty{
			"path":    {Type: "string", Description: "Path to the file to write"},
			"content": {Type: "string", Description: "Content to write to the file"},
		},
		Required: []string{"path", "content"},
	}
}

func (f *FileWrite) Execute(ctx context.Context, args map[string]any) (string, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return "", fmt.Errorf("path parameter is required and must be a non-empty string")
	}

	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("content parameter is required and must be a string")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create parent directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path), nil
}
