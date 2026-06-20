package event

import (
	"github.com/narcilee7/dew/pkg/llm"
	"github.com/narcilee7/dew/pkg/tools"
)

// Event is an agent lifecycle event.
type Event interface {
	eventMarker()
}

// TurnStartEvent marks the beginning of an agent turn.
type TurnStartEvent struct {
	Turn int `json:"turn"`
}

// TurnEndEvent marks the end of an agent turn.
type TurnEndEvent struct {
	Turn int `json:"turn"`
}

// TextDeltaEvent carries a streaming text fragment.
type TextDeltaEvent struct {
	Delta string `json:"delta"`
}

// ToolCallStartEvent marks the start of a tool call.
type ToolCallStartEvent struct {
	Call llm.ToolCall `json:"call"`
}

// ToolCallDeltaEvent carries a partial argument update.
type ToolCallDeltaEvent struct {
	CallID string `json:"call_id"`
	Delta  string `json:"delta"`
}

// ToolCallEndEvent marks the end of a tool call request.
type ToolCallEndEvent struct {
	CallID string `json:"call_id"`
}

// ToolResultEvent carries the result of a tool execution.
type ToolResultEvent struct {
	CallID string           `json:"call_id"`
	Name   string           `json:"name"`
	Result tools.ToolResult `json:"result"`
}

// AgentStartEvent marks the start of an agent run.
type AgentStartEvent struct {
	AgentID string `json:"agent_id"`
}

// AgentEndEvent marks the end of an agent run.
type AgentEndEvent struct {
	AgentID string `json:"agent_id"`
}

// ErrorEvent reports an error.
type ErrorEvent struct {
	Err error `json:"-"`
}

func (TurnStartEvent) eventMarker()     {}
func (TurnEndEvent) eventMarker()       {}
func (TextDeltaEvent) eventMarker()     {}
func (ToolCallStartEvent) eventMarker() {}
func (ToolCallDeltaEvent) eventMarker() {}
func (ToolCallEndEvent) eventMarker()   {}
func (ToolResultEvent) eventMarker()    {}
func (AgentStartEvent) eventMarker()    {}
func (AgentEndEvent) eventMarker()      {}
func (ErrorEvent) eventMarker()         {}

// EventStream is a read-only stream of events.
type EventStream interface {
	Recv() (Event, error)
	Close() error
}

// EventStreamWriter is the producer side of an event stream.
type EventStreamWriter interface {
	Send(Event) error
	Close() error
}
