package core

import (
	"context"
	"fmt"

	"github.com/narcilee7/dew/pkg/event"
	"github.com/narcilee7/dew/pkg/ai"
	"github.com/narcilee7/dew/pkg/session"
	"github.com/narcilee7/dew/pkg/tools"
)

// Hook is a single lifecycle hook.
type Hook interface {
	// Point returns the lifecycle point at which the hook fires.
	Point() HookPoint

	// Handle executes the hook. Returning a non-nil error aborts the run.
	Handle(ctx context.Context, env HookEnvironment) error
}

// HookPoint identifies a lifecycle point for hook registration.
type HookPoint int

const (
	// HookBeforeSession fires once before the agent loop starts.
	HookBeforeSession HookPoint = iota

	// HookAfterSession fires once after the agent loop ends.
	HookAfterSession

	// HookBeforeTurn fires before each LLM call.
	HookBeforeTurn

	// HookAfterTurn fires after tool results are processed for a turn.
	HookAfterTurn

	// HookBeforeToolUse fires before a tool is executed.
	HookBeforeToolUse

	// HookAfterToolUse fires after a tool returns.
	HookAfterToolUse

	// HookOnEvent fires for every event emitted by the harness.
	HookOnEvent

	// HookOnError fires when the loop encounters an error.
	HookOnError
)

// String returns a human-readable name for the hook point.
func (p HookPoint) String() string {
	switch p {
	case HookBeforeSession:
		return "before_session"
	case HookAfterSession:
		return "after_session"
	case HookBeforeTurn:
		return "before_turn"
	case HookAfterTurn:
		return "after_turn"
	case HookBeforeToolUse:
		return "before_tool_use"
	case HookAfterToolUse:
		return "after_tool_use"
	case HookOnEvent:
		return "on_event"
	case HookOnError:
		return "on_error"
	default:
		return fmt.Sprintf("hook_point_%d", p)
	}
}

// HookEnvironment carries context for a single hook invocation.
type HookEnvironment struct {
	// Harness is the harness running the agent.
	Harness Harness

	// Session is the current session.
	Session session.Session

	// Turn is the current turn number. Zero if not inside a turn.
	Turn int

	// Call is the tool call being processed, if any.
	Call *ai.ToolCall

	// Result is the tool result being processed, if any.
	Result *tools.ToolResult

	// Event is the event being emitted, if the hook point is HookOnEvent.
	Event event.Event

	// Error is the error being handled, if the hook point is HookOnError.
	Error error
}
