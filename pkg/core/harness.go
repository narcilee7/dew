package core

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/narcilee7/dew/pkg/event"
	"github.com/narcilee7/dew/pkg/ai"
	"github.com/narcilee7/dew/pkg/session"
	"github.com/narcilee7/dew/pkg/tools"
)

// Boundaries holds the isolatable resources available to a harness.
// Every capability inside the harness is accessed through these boundaries.
type Boundaries struct {
	Provider    ai.Provider
	Tools       tools.ToolRegistry
	Session     session.Store
	Logger      *slog.Logger
	ChatOptions ai.ChatOptions
}

// Harness is the core agent runtime.
// It wires together boundaries, plugins, and a loop, and exposes a lifecycle
// event stream. Upper layers (agents, skills, multi-agent patterns) build on
// top of a Harness instead of directly assembling a Runner.
type Harness interface {
	// ID returns a stable identifier for this harness instance.
	ID() string

	// Boundaries returns the isolatable resources available to the harness.
	Boundaries() Boundaries

	// Events returns a read-only stream of lifecycle events.
	Events() <-chan event.Event

	// Emit sends an event to the harness event stream.
	// It is safe to call from plugins and loops.
	Emit(ev event.Event)

	// RunHooks executes all hooks registered for the given hook point.
	RunHooks(ctx context.Context, point HookPoint, env HookEnvironment) error

	// Use registers one or more plugins with the harness.
	Use(plugins ...Plugin) error

	// Run executes the agent loop for the given session until completion or error.
	Run(ctx context.Context, sess session.Session, opts RunOptions) error
}

// Plugin participates in the agent lifecycle through hooks.
// A plugin is the primary extension mechanism for the harness.
type Plugin interface {
	// Name returns a unique, human-readable plugin name.
	Name() string

	// Hooks returns the hooks provided by this plugin.
	Hooks() []Hook
}

// DefaultHarness is the standard Harness implementation.
type DefaultHarness struct {
	id      string
	b       Boundaries
	plugins []Plugin
	loop    Loop
	events  chan event.Event
	mu      sync.RWMutex
}

// NewHarness creates a new DefaultHarness with the given boundaries.
// If id is empty, a unique id is generated.
func NewHarness(id string, b Boundaries) *DefaultHarness {
	if id == "" {
		id = fmt.Sprintf("harness-%d", time.Now().UnixNano())
	}
	if b.Logger == nil {
		b.Logger = slog.Default()
	}
	h := &DefaultHarness{
		id:     id,
		b:      b,
		events: make(chan event.Event, 256),
	}
	h.loop = &DefaultLoop{
		Safety:  &NoOpSafetyLayer{},
		Context: &SimpleContextManager{},
		Logger:  b.Logger,
	}
	return h
}

// ID returns the harness identifier.
func (h *DefaultHarness) ID() string { return h.id }

// Boundaries returns the harness boundaries.
func (h *DefaultHarness) Boundaries() Boundaries { return h.b }

// Events returns the harness event stream.
func (h *DefaultHarness) Events() <-chan event.Event { return h.events }

// Emit sends an event to the event stream.
func (h *DefaultHarness) Emit(ev event.Event) {
	if ev == nil {
		return
	}
	defer func() {
		// Ignore sends on a closed channel; the run has finished.
		_ = recover()
	}()
	select {
	case h.events <- ev:
	default:
		// Drop events if the consumer is slow to avoid blocking the loop.
		h.b.Logger.Warn("harness event channel full, dropping event", "type", fmt.Sprintf("%T", ev))
	}
}

// Use registers plugins with the harness.
func (h *DefaultHarness) Use(plugins ...Plugin) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, p := range plugins {
		if p == nil {
			return fmt.Errorf("plugin is nil")
		}
		h.plugins = append(h.plugins, p)
	}
	return nil
}

// SetLoop replaces the loop used by the harness.
// It is useful for surfaces that need to customize the default ReAct loop.
func (h *DefaultHarness) SetLoop(loop Loop) {
	if loop != nil {
		h.loop = loop
	}
}

// Loop returns the current loop. It is exposed for introspection and testing.
func (h *DefaultHarness) Loop() Loop {
	return h.loop
}

// Run executes the configured loop with plugin hooks.
// The event channel is closed when the run completes.
func (h *DefaultHarness) Run(ctx context.Context, sess session.Session, opts RunOptions) error {
	if h.b.Provider == nil {
		return fmt.Errorf("provider is nil")
	}
	if h.b.Tools == nil {
		return fmt.Errorf("tools is nil")
	}
	if sess == nil {
		return fmt.Errorf("session is nil")
	}

	defer func() {
		close(h.events)
	}()

	if err := h.RunHooks(ctx, HookBeforeSession, HookEnvironment{Harness: h, Session: sess}); err != nil {
		return err
	}

	err := h.loop.Run(ctx, h, sess, opts)
	if err != nil {
		_ = h.RunHooks(ctx, HookOnError, HookEnvironment{Harness: h, Session: sess, Error: err})
	}

	_ = h.RunHooks(ctx, HookAfterSession, HookEnvironment{Harness: h, Session: sess, Error: err})
	return err
}

// RunHooks executes all registered hooks for the given point.
func (h *DefaultHarness) RunHooks(ctx context.Context, point HookPoint, env HookEnvironment) error {
	h.mu.RLock()
	plugins := append([]Plugin(nil), h.plugins...)
	h.mu.RUnlock()

	for _, p := range plugins {
		for _, hook := range p.Hooks() {
			if hook == nil || hook.Point() != point {
				continue
			}
			if err := hook.Handle(ctx, env); err != nil {
				return fmt.Errorf("plugin %q hook %s: %w", p.Name(), point.String(), err)
			}
		}
	}
	return nil
}
