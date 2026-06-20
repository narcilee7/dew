package agent

import (
	"context"
	"time"

	"github.com/narcilee7/dew/pkg/core"
	"github.com/narcilee7/dew/pkg/event"
	"github.com/narcilee7/dew/pkg/llm"
)

// Capability describes what an agent can do.
type Capability struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tools       []string `json:"tools"`
	Sandbox     string   `json:"sandbox"`
	Model       string   `json:"model"`
}

// AgentInfo describes a running agent process.
type AgentInfo struct {
	ID           string            `json:"id"`
	Address      string            `json:"address"` // gRPC address
	Capabilities []Capability      `json:"capabilities"`
	MCPEndpoint  string            `json:"mcp_endpoint,omitempty"`
	A2AEndpoint  string            `json:"a2a_endpoint,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	StartedAt    time.Time         `json:"started_at"`
}

// Task is the unit of delegation between agents.
type Task struct {
	ID       string        `json:"id"`
	ParentID string        `json:"parent_id"`
	Goal     string        `json:"goal"`
	Context  []llm.Message `json:"context,omitempty"`
	Spec     AgentSpec     `json:"spec"`
	MaxTurns int           `json:"max_turns"`
	Timeout  time.Duration `json:"timeout"`
}

// TaskResult is the outcome of a delegated task.
type TaskResult struct {
	TaskID    string       `json:"task_id"`
	Status    core.Status  `json:"status"`
	Summary   string       `json:"summary"`
	Artifacts []Artifact   `json:"artifacts,omitempty"`
	Events    []event.Event `json:"events,omitempty"`
	Error     string       `json:"error,omitempty"`
}

// Artifact is a file or output produced by an agent.
type Artifact struct {
	Name    string `json:"name"`
	Type    string `json:"type"` // "file", "diff", "log", "test_result"
	Path    string `json:"path,omitempty"`
	Content []byte `json:"content,omitempty"`
}

// AgentSpec describes the desired shape of an agent.
type AgentSpec struct {
	Capability string            `json:"capability"`
	Sandbox    string            `json:"sandbox,omitempty"`
	Model      string            `json:"model,omitempty"`
	Tools      []string          `json:"tools,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// Agent is a runnable dew instance with an exposed gRPC service.
type Agent interface {
	ID() string
	Info() AgentInfo
	Capabilities() []Capability
	Run(ctx context.Context, task Task, events chan<- event.Event) (TaskResult, error)
	Spawn(ctx context.Context, task Task) (Handle, error)
	Cancel(ctx context.Context) error
}

// Handle is a reference to a running child agent.
type Handle interface {
	ID() string
	Wait(ctx context.Context) (TaskResult, error)
	Cancel(ctx context.Context) error
}

// Registry is optional dynamic agent discovery.
type Registry interface {
	Register(ctx context.Context, info AgentInfo) error
	Deregister(ctx context.Context, id string) error
	List(ctx context.Context, filter CapabilityFilter) ([]AgentInfo, error)
	Resolve(ctx context.Context, id string) (AgentInfo, error)
}

// CapabilityFilter filters agents by capability.
type CapabilityFilter struct {
	Name  string
	Tools []string
}

// Pool manages pre-warmed agent processes.
type Pool interface {
	Acquire(ctx context.Context, spec AgentSpec) (Agent, error)
	Release(ctx context.Context, agent Agent) error
	Warm(ctx context.Context, spec AgentSpec, count int) error
	Stop(ctx context.Context) error
}
