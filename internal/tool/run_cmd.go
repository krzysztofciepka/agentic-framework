package tool

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type RunCmd struct{}

func (r *RunCmd) Name() string        { return "run_cmd" }
func (r *RunCmd) Description() string { return "Execute a shell command in a sandboxed environment" }
func (r *RunCmd) Category() string    { return "system" }

func (r *RunCmd) Parameters() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]ParamProperty{
			"command": {Type: "string", Description: "The shell command to execute"},
			"workdir": {Type: "string", Description: "Working directory for the command"},
		},
		Required: []string{"command"},
	}
}

func (r *RunCmd) Execute(ctx context.Context, args map[string]any) (string, error) {
	command, ok := args["command"].(string)
	if !ok || command == "" {
		return "", fmt.Errorf("command parameter is required and must be a non-empty string")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)

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
