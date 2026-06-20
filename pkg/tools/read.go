package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

// ReadTool reads files from the filesystem.
type ReadTool struct{}

// ReadArgs is the input schema for the read tool.
type ReadArgs struct {
	Path string `json:"path"`
}

// Name returns the tool name.
func (t *ReadTool) Name() string { return "read" }

// Description returns the tool description.
func (t *ReadTool) Description() string {
	return "Read the contents of a file at the given path."
}

// Schema returns the tool's JSON schema.
func (t *ReadTool) Schema() ToolSchema {
	return ToolSchema{
		Type:        "object",
		Description: t.Description(),
		Properties: map[string]Property{
			"path": {
				Type:        "string",
				Description: "Relative path to the file to read",
			},
		},
		Required: []string{"path"},
	}
}

// Execute reads a file.
func (t *ReadTool) Execute(ctx context.Context, env ToolEnvironment, args json.RawMessage) (ToolResult, error) {
	var ra ReadArgs
	if err := json.Unmarshal(args, &ra); err != nil {
		return ToolResult{}, fmt.Errorf("invalid read args: %w", err)
	}
	if ra.Path == "" {
		return ToolResult{}, fmt.Errorf("path is required")
	}

	data, err := env.FileSystem().Read(ctx, ra.Path)
	if err != nil {
		return ToolResult{}, err
	}

	return ToolResult{
		Content: []ContentPart{{Type: "text", Text: string(data)}},
		Metadata: ToolResultMetadata{
			Status: ToolExecuteStatusSuccess,
		},
	}, nil
}
