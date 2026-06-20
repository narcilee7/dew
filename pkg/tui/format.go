package tui

import (
	"fmt"

	"github.com/narcilee7/dew/pkg/event"
)

// formatEvent converts a lifecycle event into a short human-readable line.
func formatEvent(ev event.Event) string {
	switch e := ev.(type) {
	case event.AgentStartEvent:
		return "▶ run started"
	case event.AgentEndEvent:
		return "■ run ended"
	case event.TurnStartEvent:
		return fmt.Sprintf("turn %d", e.Turn)
	case event.TurnEndEvent:
		return fmt.Sprintf("/turn %d", e.Turn)
	case event.TextDeltaEvent:
		return e.Delta
	case event.ToolCallStartEvent:
		return fmt.Sprintf("→ %s(%s)", e.Call.Name, e.Call.ID)
	case event.ToolCallDeltaEvent:
		return e.Delta
	case event.ToolCallEndEvent:
		return fmt.Sprintf("← %s", e.CallID)
	case event.ToolResultEvent:
		status := "ok"
		if e.Result.Metadata.Status != "" {
			status = string(e.Result.Metadata.Status)
		}
		if e.Result.Metadata.Error != "" {
			status = "error: " + e.Result.Metadata.Error
		}
		return fmt.Sprintf("result %s: %s", e.Name, status)
	case event.ErrorEvent:
		return "error: " + e.Err.Error()
	default:
		return fmt.Sprintf("%T", ev)
	}
}
