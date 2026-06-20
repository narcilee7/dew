package plan

import (
	"context"
	"time"
)

// Step is a single unit of work within a plan.
type Step struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	DependsOn   []string `json:"depends_on,omitempty"`
}

// Plan is a high-level task decomposition.
type Plan struct {
	ID        string    `json:"id"`
	Goal      string    `json:"goal"`
	SessionID string    `json:"session_id"`
	Steps     []Step    `json:"steps"`
	CreatedAt time.Time `json:"created_at"`
}

// Snapshot captures filesystem and session state at a point in time.
type Snapshot struct {
	Files     map[string]string `json:"files,omitempty"`     // path -> content hash
	Messages  []byte            `json:"messages,omitempty"`  // serialized session state
	EventIndex int              `json:"event_index"`
}

// Checkpoint captures state before a risky operation.
type Checkpoint struct {
	ID        string    `json:"id"`
	PlanID    string    `json:"plan_id"`
	StepID    string    `json:"step_id,omitempty"`
	Label     string    `json:"label"`
	Snapshot  Snapshot  `json:"snapshot"`
	CreatedAt time.Time `json:"created_at"`
}

// CheckpointInfo is a lightweight checkpoint reference.
type CheckpointInfo struct {
	ID        string    `json:"id"`
	Label     string    `json:"label"`
	CreatedAt time.Time `json:"created_at"`
}

// Store persists plans and checkpoints.
type Store interface {
	Create(ctx context.Context, plan Plan) error
	Load(ctx context.Context, id string) (Plan, error)
	SaveCheckpoint(ctx context.Context, planID string, cp Checkpoint) error
	LoadCheckpoint(ctx context.Context, planID string, checkpointID string) (Checkpoint, error)
	ListCheckpoints(ctx context.Context, planID string) ([]CheckpointInfo, error)
}
