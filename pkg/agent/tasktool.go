package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/narcilee7/dew/pkg/tools"
)

// TaskTool delegates work to another agent.
type TaskTool struct {
	Pool Pool
}

// TaskArgs is the input schema for the task tool.
type TaskArgs struct {
	Goal       string   `json:"goal"`
	Capability string   `json:"capability"`
	Tools      []string `json:"tools,omitempty"`
	MaxTurns   int      `json:"max_turns,omitempty"`
	Timeout    int64    `json:"timeout_ms,omitempty"`
}

// Name returns the tool name.
func (t *TaskTool) Name() string { return "task" }

// Description returns the tool description.
func (t *TaskTool) Description() string {
	return "Delegate a subtask to another agent with a specific capability."
}

// Schema returns the tool's JSON schema.
func (t *TaskTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type:        "object",
		Description: t.Description(),
		Properties: map[string]tools.Property{
			"goal": {
				Type:        "string",
				Description: "Clear objective for the subagent",
			},
			"capability": {
				Type:        "string",
				Description: "Required agent capability, e.g. coder, reviewer, tester",
			},
			"tools": {
				Type:        "array",
				Description: "Allowed tools for the subagent",
			},
			"max_turns": {
				Type:        "integer",
				Description: "Maximum agent turns",
			},
			"timeout_ms": {
				Type:        "integer",
				Description: "Timeout in milliseconds",
			},
		},
		Required: []string{"goal", "capability"},
	}
}

// Execute delegates to another agent.
func (t *TaskTool) Execute(ctx context.Context, env tools.ToolEnvironment, args json.RawMessage) (tools.ToolResult, error) {
	if t.Pool == nil {
		return tools.ToolResult{}, fmt.Errorf("agent pool not configured")
	}

	var ta TaskArgs
	if err := json.Unmarshal(args, &ta); err != nil {
		return tools.ToolResult{}, fmt.Errorf("invalid task args: %w", err)
	}

	spec := AgentSpec{
		Capability: ta.Capability,
		Tools:      ta.Tools,
	}

	worker, err := t.Pool.Acquire(ctx, spec)
	if err != nil {
		return tools.ToolResult{}, fmt.Errorf("acquire agent: %w", err)
	}
	defer t.Pool.Release(ctx, worker)

	task := Task{
		Goal:     ta.Goal,
		Spec:     spec,
		MaxTurns: ta.MaxTurns,
		Timeout:  time.Duration(ta.Timeout) * time.Millisecond,
	}

	result, err := worker.Run(ctx, task, nil)
	if err != nil {
		return tools.ToolResult{}, err
	}

	if result.Error != "" {
		return tools.ToolResult{
			Content: []tools.ContentPart{{Type: "text", Text: result.Summary}},
			Metadata: tools.ToolResultMetadata{
				Error:  result.Error,
				Status: tools.ToolExecuteStatusError,
			},
		}, nil
	}

	return tools.ToolResult{
		Content: []tools.ContentPart{{Type: "text", Text: result.Summary}},
		Metadata: tools.ToolResultMetadata{
			Status: tools.ToolExecuteStatusSuccess,
		},
	}, nil
}
