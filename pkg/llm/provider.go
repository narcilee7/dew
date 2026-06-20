package llm

import (
	"context"
	"encoding/json"
)

// Model identifies a specific LLM and its provider.
type Model struct {
	ID       string
	Provider string
	API      string
}

// Message is a single message in a conversation transcript.
// It is the canonical definition shared across session, core, and provider layers.
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

// ToolCall represents a single tool invocation requested by the model.
type ToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// Tool describes a callable tool in the LLM context.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

// Context is the input to an LLM call.
type Context struct {
	SystemPrompt string
	Messages     []Message
	Tools        []Tool
}

// Response is a normalized LLM response.
type Response struct {
	Content   string
	ToolCalls []ToolCall
}

// Options configures a streaming or completion request.
type Options struct {
	APIKey      string
	BaseURL     string
	MaxTokens   int
	Temperature float64
}

// Event represents a normalized streaming event.
type Event interface {
	eventMarker()
}

// TextStartEvent marks the start of text generation.
type TextStartEvent struct{}

// TextDeltaEvent carries a streaming text fragment.
type TextDeltaEvent struct{ Delta string }

// TextEndEvent marks the end of text generation.
type TextEndEvent struct{}

// ToolCallStartEvent marks the start of a tool call.
type ToolCallStartEvent struct {
	ID   string
	Name string
}

// ToolCallDeltaEvent carries a partial arguments update.
type ToolCallDeltaEvent struct {
	ID        string
	Arguments string
}

// ToolCallEndEvent marks the end of a tool call.
type ToolCallEndEvent struct{ ID string }

// DoneEvent marks stream completion.
type DoneEvent struct{}

// ErrorEvent reports a stream error.
type ErrorEvent struct{ Err error }

func (TextStartEvent) eventMarker()     {}
func (TextDeltaEvent) eventMarker()     {}
func (TextEndEvent) eventMarker()       {}
func (ToolCallStartEvent) eventMarker() {}
func (ToolCallDeltaEvent) eventMarker() {}
func (ToolCallEndEvent) eventMarker()   {}
func (DoneEvent) eventMarker()          {}
func (ErrorEvent) eventMarker()         {}

// EventStream consumes normalized LLM streaming events.
type EventStream interface {
	Recv() (Event, error)
	Close() error
}

// Provider is the unified interface for all LLM backends.
type Provider interface {
	Stream(ctx context.Context, model Model, context Context, opts Options) (EventStream, error)
	Complete(ctx context.Context, model Model, context Context, opts Options) (Response, error)
}
