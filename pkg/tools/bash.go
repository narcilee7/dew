package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/narcilee7/dew/pkg/sandbox"
)

// BashTool executes shell commands in the sandbox.
type BashTool struct{}

// BashArgs is the input schema for the bash tool.
type BashArgs struct {
	Command string `json:"command"`
	Timeout int64  `json:"timeout,omitempty"`
}

// Name returns the tool name.
func (t *BashTool) Name() string { return "bash" }

// Description returns the tool description.
func (t *BashTool) Description() string {
	return "Execute a shell command in the working directory."
}

// Schema returns the tool's JSON schema.
func (t *BashTool) Schema() ToolSchema {
	return ToolSchema{
		Type:        "object",
		Description: t.Description(),
		Properties: map[string]Property{
			"command": {
				Type:        "string",
				Description: "Shell command to execute",
			},
			"timeout": {
				Type:        "integer",
				Description: "Timeout in milliseconds (optional)",
			},
		},
		Required: []string{"command"},
	}
}

// Execute runs a shell command.
func (t *BashTool) Execute(ctx context.Context, env ToolEnvironment, args json.RawMessage) (ToolResult, error) {
	var ba BashArgs
	if err := json.Unmarshal(args, &ba); err != nil {
		return ToolResult{}, fmt.Errorf("invalid bash args: %w", err)
	}
	if ba.Command == "" {
		return ToolResult{}, fmt.Errorf("command is required")
	}

	cmd := sandbox.Command{
		Args:    []string{"sh", "-c", ba.Command},
		WorkDir: "",
	}
	if ba.Timeout > 0 {
		cmd.Timeout = time.Duration(ba.Timeout) * time.Millisecond
	}

	res, err := env.Sandbox().Exec(ctx, cmd)
	if err != nil {
		return ToolResult{}, err
	}

	output := string(res.Stdout)
	if len(res.Stderr) > 0 {
		output += "\n[stderr]\n" + string(res.Stderr)
	}
	if res.ExitCode != 0 {
		return ToolResult{
			Content: []ContentPart{{Type: "text", Text: output}},
			Metadata: ToolResultMetadata{
				Error:  fmt.Sprintf("exit code %d", res.ExitCode),
				Status: ToolExecuteStatusError,
			},
		}, nil
	}

	return ToolResult{
		Content: []ContentPart{{Type: "text", Text: output}},
		Metadata: ToolResultMetadata{
			Status: ToolExecuteStatusSuccess,
		},
	}, nil
}
