package tool

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Execute struct{}

func (e *Execute) Name() string        { return "execute" }
func (e *Execute) Description() string { return "Run a binary with arguments in a sandboxed environment" }
func (e *Execute) Category() string    { return "system" }

func (e *Execute) Parameters() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]ParamProperty{
			"binary":  {Type: "string", Description: "Binary to execute"},
			"args":    {Type: "string", Description: "Space-separated arguments"},
			"workdir": {Type: "string", Description: "Working directory"},
		},
		Required: []string{"binary"},
	}
}

func (e *Execute) Execute(ctx context.Context, args map[string]any) (string, error) {
	binary, ok := args["binary"].(string)
	if !ok || binary == "" {
		return "", fmt.Errorf("binary parameter is required and must be a non-empty string")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var cmdArgs []string
	if rawArgs, ok := args["args"].(string); ok && rawArgs != "" {
		cmdArgs = strings.Fields(rawArgs)
	}

	cmd := exec.CommandContext(ctx, binary, cmdArgs...)

	if workdir, ok := args["workdir"].(string); ok && workdir != "" {
		cmd.Dir = workdir
	}

	const maxOutput = 1 << 20
	output, err := cmd.CombinedOutput()
	if len(output) > maxOutput {
		output = output[:maxOutput]
	}

	if err != nil {
		return fmt.Sprintf("Error: %v\n%s", err, string(output)), nil
	}
	return string(output), nil
}
