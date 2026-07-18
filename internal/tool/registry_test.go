package tool

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRegistry(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&WebSearch{})

	tool, err := reg.Get("web_search")
	if err != nil {
		t.Fatalf("expected to get web_search, got error: %v", err)
	}

	if tool.Name() != "web_search" {
		t.Errorf("expected Name() = %q, got %q", "web_search", tool.Name())
	}

	_, err = reg.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent tool, got nil")
	}
}

func TestDefaultRegistry(t *testing.T) {
	reg := DefaultRegistry()
	all := reg.All()

	if len(all) != 6 {
		t.Errorf("expected 6 tools, got %d", len(all))
	}

	names := make(map[string]bool)
	for _, tool := range all {
		names[tool.Name()] = true
	}
	expected := []string{"web_search", "web_fetch", "run_cmd", "file_system_read", "file_system_write", "execute"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected tool %q in default registry", name)
		}
	}
}

func TestRunCmd(t *testing.T) {
	r := &RunCmd{}
	result, err := r.Execute(context.Background(), map[string]any{
		"command": "echo hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello\n" {
		t.Errorf("expected %q, got %q", "hello\n", result)
	}
}

func TestFileReadWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := "hello world"

	w := &FileWrite{}
	result, err := w.Execute(context.Background(), map[string]any{
		"path":    path,
		"content": content,
	})
	if err != nil {
		t.Fatalf("write error: %v", err)
	}
	if result == "" {
		t.Fatal("expected non-empty write result")
	}

	r := &FileRead{}
	result, err = r.Execute(context.Background(), map[string]any{
		"path": path,
	})
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if result != content {
		t.Errorf("expected %q, got %q", content, result)
	}
}

func TestExecute(t *testing.T) {
	e := &Execute{}
	result, err := e.Execute(context.Background(), map[string]any{
		"binary": "echo",
		"args":   "hello world",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello world\n" {
		t.Errorf("expected %q, got %q", "hello world\n", result)
	}
}

func TestWebFetch(t *testing.T) {
	w := &WebFetch{}
	result, err := w.Execute(context.Background(), map[string]any{
		"url": "https://httpbin.org/get",
	})
	if err != nil {
		t.Skipf("skipping web_fetch test (network unavailable): %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result from web_fetch")
	}
}

func TestWebSearchEnvMissing(t *testing.T) {
	os.Unsetenv("BRAVE_API_KEY")
	w := &WebSearch{}
	_, err := w.Execute(context.Background(), map[string]any{
		"query": "test",
	})
	if err == nil {
		t.Fatal("expected error when BRAVE_API_KEY is not set")
	}
}

func TestFileReadErrors(t *testing.T) {
	r := &FileRead{}
	result, err := r.Execute(context.Background(), map[string]any{
		"path": "/nonexistent/path/file.txt",
	})
	if err != nil {
		t.Fatalf("file_read should return error content, not error: %v", err)
	}
	if result == "" {
		t.Fatal("expected error message in content")
	}
}

func TestFileWriteNested(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c.txt")
	content := "nested"

	w := &FileWrite{}
	_, err := w.Execute(context.Background(), map[string]any{
		"path":    path,
		"content": content,
	})
	if err != nil {
		t.Fatalf("write error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if string(data) != content {
		t.Errorf("expected %q, got %q", content, string(data))
	}
}
