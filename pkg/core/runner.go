package core

import (
	"context"
	"log/slog"

	"github.com/narcilee7/dew/pkg/event"
	"github.com/narcilee7/dew/pkg/llm"
	"github.com/narcilee7/dew/pkg/session"
	"github.com/narcilee7/dew/pkg/tools"
)

// SafetyLayer decides whether tool calls are allowed.
type SafetyLayer interface {
	BeforeToolCall(ctx context.Context, call llm.ToolCall, env tools.ToolEnvironment) (bool, error)
	AfterToolCall(ctx context.Context, call llm.ToolCall, result tools.ToolResult) (tools.ToolResult, error)
}

// NoOpSafetyLayer allows all tool calls.
type NoOpSafetyLayer struct{}

func (n *NoOpSafetyLayer) BeforeToolCall(ctx context.Context, call llm.ToolCall, env tools.ToolEnvironment) (bool, error) {
	return true, nil
}

func (n *NoOpSafetyLayer) AfterToolCall(ctx context.Context, call llm.ToolCall, result tools.ToolResult) (tools.ToolResult, error) {
	return result, nil
}

// ContextManager builds the LLM context from a session.
type ContextManager interface {
	Build(sess session.Session, tools tools.ToolRegistry) (llm.Context, error)
}

// SimpleContextManager builds context from session messages and tool schemas.
type SimpleContextManager struct {
	SystemPrompt string
}

// Build constructs an llm.Context.
func (s *SimpleContextManager) Build(sess session.Session, registry tools.ToolRegistry) (llm.Context, error) {
	var llmTools []llm.Tool
	for _, t := range registry.List() {
		schema := t.Schema()
		llmTools = append(llmTools, llm.Tool{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters:  schema,
		})
	}

	system := s.SystemPrompt
	if system == "" {
		system = "You are dew, a helpful coding agent."
	}

	return llm.Context{
		SystemPrompt: system,
		Messages:     sess.Messages(),
		Tools:        llmTools,
	}, nil
}

// Runner executes the agent loop for a session.
//
// Deprecated: Runner is a compatibility wrapper around Harness. New code should
// create a Harness directly and use its plugin hook system.
type Runner struct {
	Provider llm.Provider
	Tools    tools.ToolRegistry
	Safety   SafetyLayer
	Context  ContextManager
	Logger   *slog.Logger
}

// NewRunner creates a new Runner with defaults.
//
// Deprecated: use NewHarness instead.
func NewRunner(provider llm.Provider, tools tools.ToolRegistry) *Runner {
	logger := slog.Default()
	return &Runner{
		Provider: provider,
		Tools:    tools,
		Safety:   &NoOpSafetyLayer{},
		Context:  &SimpleContextManager{},
		Logger:   logger,
	}
}

// Run executes the agent loop until completion or error.
// Events are emitted on the provided channel.
//
// Deprecated: use Harness.Run instead.
func (r *Runner) Run(ctx context.Context, sess session.Session, opts RunOptions, events chan<- event.Event) error {
	b := Boundaries{
		Provider: r.Provider,
		Tools:    r.Tools,
		Session:  nil,
		Logger:   r.Logger,
	}
	h := NewHarness("", b)

	// Replace the default loop with one that uses Runner's safety/context.
	h.loop = &DefaultLoop{
		Safety:  r.Safety,
		Context: r.Context,
		Logger:  r.Logger,
	}

	// Forward harness events to the caller's channel.
	go func() {
		for ev := range h.Events() {
			select {
			case events <- ev:
			case <-ctx.Done():
				return
			}
		}
	}()

	return h.Run(ctx, sess, opts)
}
