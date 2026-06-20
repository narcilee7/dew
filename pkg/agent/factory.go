package agent

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/narcilee7/dew/pkg/core"
	"github.com/narcilee7/dew/pkg/ai"
	"github.com/narcilee7/dew/pkg/session"
	"github.com/narcilee7/dew/pkg/tools"
)

var factoryCounter atomic.Int64

// LocalFactory creates in-process LocalAgent instances backed by core.Harness.
type LocalFactory struct {
	Provider   ai.Provider
	Session    session.Store
	Registry   tools.ToolRegistry
	Logger     *slog.Logger
	Capability map[string]Capability
}

// NewLocalFactory creates a new LocalFactory.
func NewLocalFactory(provider ai.Provider, store session.Store, registry tools.ToolRegistry, logger *slog.Logger) *LocalFactory {
	return &LocalFactory{
		Provider: provider,
		Session:  store,
		Registry: registry,
		Logger:   logger,
		Capability: map[string]Capability{
			"coder": {
				Name:        "coder",
				Description: "General coding assistant",
				Tools:       []string{"read", "bash", "task"},
				Sandbox:     "local",
				Model:       "openai:gpt-4o",
			},
			"reviewer": {
				Name:        "reviewer",
				Description: "Code reviewer",
				Tools:       []string{"read", "bash"},
				Sandbox:     "local",
				Model:       "openai:gpt-4o",
			},
			"tester": {
				Name:        "tester",
				Description: "Test engineer",
				Tools:       []string{"read", "bash"},
				Sandbox:     "local",
				Model:       "openai:gpt-4o",
			},
		},
	}
}

// Create builds a LocalAgent for the given spec.
func (f *LocalFactory) Create(ctx context.Context, spec AgentSpec) (Agent, error) {
	cap, ok := f.Capability[spec.Capability]
	if !ok {
		return nil, fmt.Errorf("unknown capability %q", spec.Capability)
	}
	if spec.Sandbox != "" {
		cap.Sandbox = spec.Sandbox
	}
	if spec.Model != "" {
		cap.Model = spec.Model
	}
	if len(spec.Tools) > 0 {
		cap.Tools = spec.Tools
	}

	filtered := f.Registry.Filter(cap.Tools)

	boundaries := core.Boundaries{
		Provider: f.Provider,
		Tools:    filtered,
		Session:  f.Session,
		Logger:   f.Logger,
	}

	id := fmt.Sprintf("agent-%d-%d", time.Now().UnixNano(), factoryCounter.Add(1))
	info := AgentInfo{
		ID:           id,
		Capabilities: []Capability{cap},
		Metadata:     spec.Metadata,
	}

	return NewLocalAgent(id, info, boundaries, f.Session, filtered), nil
}
