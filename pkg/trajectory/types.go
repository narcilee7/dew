package trajectory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/narcilee7/dew/pkg/event"
)

// Trajectory is a persistent record of an agent session.
type Trajectory struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Goal      string    `json:"goal,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EventRecord wraps a lifecycle event with metadata for storage and replay.
// Payload is stored as raw JSON; call DecodePayload to obtain a typed event.
type EventRecord struct {
	Seq       int             `json:"seq"`
	Timestamp time.Time       `json:"timestamp"`
	SessionID string          `json:"session_id,omitempty"`
	Turn      int             `json:"turn,omitempty"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
}

// NewEventRecord builds a record for the given event.
func NewEventRecord(ev event.Event, seq int, sessionID string, turn int) (EventRecord, error) {
	raw, err := json.Marshal(ev)
	if err != nil {
		return EventRecord{}, err
	}
	return EventRecord{
		Seq:       seq,
		Timestamp: time.Now(),
		SessionID: sessionID,
		Turn:      turn,
		Type:      eventTypeName(ev),
		Payload:   raw,
	}, nil
}

// DecodePayload returns the concrete event type stored in Payload.
func (r *EventRecord) DecodePayload() (event.Event, error) {
	return decodeEvent(r.Type, r.Payload)
}

// Store persists trajectories and their event records.
type Store interface {
	Create(ctx context.Context, traj Trajectory) error
	Save(ctx context.Context, traj Trajectory) error
	Load(ctx context.Context, id string) (Trajectory, error)
	AppendEvent(ctx context.Context, trajectoryID string, rec EventRecord) error
	ListEvents(ctx context.Context, trajectoryID string) ([]EventRecord, error)
}

// eventTypeName returns the concrete event type name.
func eventTypeName(ev event.Event) string {
	return fmt.Sprintf("%T", ev)
}

// decodeEvent reconstructs a concrete event from its type name and raw payload.
func decodeEvent(typeName string, data []byte) (event.Event, error) {
	var ev event.Event
	switch typeName {
	case "event.AgentStartEvent":
		var v event.AgentStartEvent
		ev = &v
	case "event.AgentEndEvent":
		var v event.AgentEndEvent
		ev = &v
	case "event.TurnStartEvent":
		var v event.TurnStartEvent
		ev = &v
	case "event.TurnEndEvent":
		var v event.TurnEndEvent
		ev = &v
	case "event.TextDeltaEvent":
		var v event.TextDeltaEvent
		ev = &v
	case "event.ToolCallStartEvent":
		var v event.ToolCallStartEvent
		ev = &v
	case "event.ToolCallDeltaEvent":
		var v event.ToolCallDeltaEvent
		ev = &v
	case "event.ToolCallEndEvent":
		var v event.ToolCallEndEvent
		ev = &v
	case "event.ToolResultEvent":
		var v event.ToolResultEvent
		ev = &v
	case "event.ErrorEvent":
		var v event.ErrorEvent
		ev = &v
	default:
		return nil, fmt.Errorf("unknown event type: %s", typeName)
	}
	if err := json.Unmarshal(data, ev); err != nil {
		return nil, err
	}
	return ev, nil
}
