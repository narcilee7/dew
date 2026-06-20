package log

import (
	"context"
	"encoding/json"
	"time"

	"github.com/narcilee7/dew/pkg/llm"
)

// ToolCallRecord records a tool invocation.
type ToolCallRecord struct {
	Timestamp  time.Time       `json:"timestamp"`
	SessionID  string          `json:"session_id"`
	AgentID    string          `json:"agent_id,omitempty"`
	CallID     string          `json:"call_id"`
	ToolName   string          `json:"tool_name"`
	Arguments  json.RawMessage `json:"arguments"`
	Status     string          `json:"status"`
	Error      string          `json:"error,omitempty"`
	ExitCode   int             `json:"exit_code,omitempty"`
	DurationMs int64           `json:"duration_ms,omitempty"`
}

// SafetyRecord records a safety decision.
type SafetyRecord struct {
	Timestamp time.Time `json:"timestamp"`
	SessionID string    `json:"session_id"`
	CallID    string    `json:"call_id"`
	Decision  string    `json:"decision"` // allow, ask, deny
	Reason    string    `json:"reason,omitempty"`
	Rule      string    `json:"rule,omitempty"`
}

// DelegationRecord records a subagent delegation.
type DelegationRecord struct {
	Timestamp   time.Time `json:"timestamp"`
	ParentID    string    `json:"parent_id"`
	ChildID     string    `json:"child_id"`
	Capability  string    `json:"capability"`
	GoalSummary string    `json:"goal_summary"`
}

// FileAccessRecord records a filesystem access.
type FileAccessRecord struct {
	Timestamp time.Time `json:"timestamp"`
	SessionID string    `json:"session_id"`
	Operation string    `json:"operation"` // read, write, list, exists
	Path      string    `json:"path"`
	Allowed   bool      `json:"allowed"`
}

// ToolCallStart builds a record for the start of a tool call.
func ToolCallStart(sessionID string, call llm.ToolCall) ToolCallRecord {
	return ToolCallRecord{
		Timestamp: time.Now(),
		SessionID: sessionID,
		CallID:    call.ID,
		ToolName:  call.Name,
		Arguments: call.Arguments,
		Status:    "started",
	}
}

// ToolCallFinish builds a record for the completion of a tool call.
func ToolCallFinish(base ToolCallRecord, err error) ToolCallRecord {
	base.Timestamp = time.Now()
	if err != nil {
		base.Status = "error"
		base.Error = err.Error()
	} else {
		base.Status = "success"
	}
	return base
}

// AuditLogger records security-relevant and action-relevant events.
type AuditLogger interface {
	ToolCall(ctx context.Context, record ToolCallRecord) error
	SafetyDecision(ctx context.Context, record SafetyRecord) error
	Delegation(ctx context.Context, record DelegationRecord) error
	FileAccess(ctx context.Context, record FileAccessRecord) error
}
