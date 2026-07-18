package tool

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var allowedDir = resolveAllowedDir()

func resolveAllowedDir() string {
	if d := os.Getenv("ALLOWED_DIR"); d != "" {
		return filepath.Clean(d)
	}
	return filepath.Clean("/tmp/agentic-framework")
}

func safeResolve(baseDir, userPath string) (string, error) {
	cleaned := filepath.Clean(userPath)
	fullPath := filepath.Join(baseDir, cleaned)
	fullPath = filepath.Clean(fullPath)
	baseDir = filepath.Clean(baseDir)
	if !strings.HasPrefix(fullPath, baseDir+string(filepath.Separator)) && fullPath != baseDir {
		return "", fmt.Errorf("path escapes allowed directory: %s", userPath)
	}
	return fullPath, nil
}

type FileRead struct{}

func (f *FileRead) Name() string        { return "file_system_read" }
func (f *FileRead) Description() string { return "Read contents of a file" }
func (f *FileRead) Category() string    { return "file" }

func (f *FileRead) Parameters() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]ParamProperty{
			"path": {Type: "string", Description: "Path to the file to read"},
		},
		Required: []string{"path"},
	}
}

func (f *FileRead) Execute(ctx context.Context, args map[string]any) (string, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return "", fmt.Errorf("path parameter is required and must be a non-empty string")
	}

	resolved, err := safeResolve(allowedDir, path)
	if err != nil {
		return "", err
	}

	if err := ctx.Err(); err != nil {
		return "", err
	}

	data, err := os.ReadFile(resolved)
	if err != nil {
		return fmt.Sprintf("Error reading file: %v", err), nil
	}

	const maxSize = 1 << 20
	if len(data) > maxSize {
		data = data[:maxSize]
	}

	return string(data), nil
}
