package core

import "time"

// Re-export role constants for convenience.
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
	RoleTool      = "tool"
)

// RunOptions configures a single agent run.
type RunOptions struct {
	MaxTurns     int
	Timeout      time.Duration
	Model        string
	SystemPrompt string
}

// Status represents the outcome status of a task or run.
type Status string

const (
	StatusSuccess   Status = "success"
	StatusFailure   Status = "failure"
	StatusCancelled Status = "cancelled"
	StatusRunning   Status = "running"
)
