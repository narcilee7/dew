package tools

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/narcilee7/dew/pkg/fs"
	"github.com/narcilee7/dew/pkg/sandbox"
)

// Tool is the interface for all tools available to an agent.
type Tool interface {
	Name() string
	Description() string
	Schema() ToolSchema
	Execute(ctx context.Context, env ToolEnvironment, args json.RawMessage) (ToolResult, error)
}

// ToolEnvironment is the execution context passed to every tool.
type ToolEnvironment interface {
	FileSystem() fs.FileSystem
	Sandbox() sandbox.Sandbox
	Logger() *slog.Logger
}

// ToolRegistry holds and dispatches tools.
type ToolRegistry interface {
	Register(t Tool) error
	Get(name string) (Tool, bool)
	List() []Tool
	Filter(names []string) ToolRegistry
}

// ToolSchema describes tool parameters as JSON schema.
type ToolSchema struct {
	Type        string              `json:"type"`
	Properties  map[string]Property `json:"properties,omitempty"`
	Required    []string            `json:"required,omitempty"`
	Description string              `json:"description,omitempty"`
}

// Property is a JSON schema property.
type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

// ToolResult is the outcome of a tool execution.
type ToolResult struct {
	Content  []ContentPart      `json:"content"`
	Metadata ToolResultMetadata `json:"metadata"`
}

// ToolResultMetadata carries execution status and optional error message.
type ToolResultMetadata struct {
	Status ToolExecuteStatus `json:"status"`
	Error  string            `json:"error,omitempty"`
}

// ContentPart is a fragment of tool result content.
type ContentPart struct {
	Type string `json:"type"` // "text" or "image"
	Text string `json:"text,omitempty"`
	Data []byte `json:"data,omitempty"`
}

// ToolExecuteStatus represents the lifecycle status of a tool execution.
type ToolExecuteStatus string

const (
	ToolExecuteStatusPending  ToolExecuteStatus = "pending"
	ToolExecuteStatusRunning  ToolExecuteStatus = "running"
	ToolExecuteStatusComplete ToolExecuteStatus = "complete"
	ToolExecuteStatusSuccess  ToolExecuteStatus = "success"
	ToolExecuteStatusError    ToolExecuteStatus = "error"
	ToolExecuteStatusTimeout  ToolExecuteStatus = "timeout"
	ToolExecuteStatusCanceled ToolExecuteStatus = "canceled"
)
