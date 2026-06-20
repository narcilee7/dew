package core

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/narcilee7/dew/pkg/event"
	"github.com/narcilee7/dew/pkg/fs"
	"github.com/narcilee7/dew/pkg/llm"
	"github.com/narcilee7/dew/pkg/sandbox"
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
type Runner struct {
	Provider llm.Provider
	Tools    tools.ToolRegistry
	Safety   SafetyLayer
	Context  ContextManager
	Logger   *slog.Logger
}

// NewRunner creates a new Runner with defaults.
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
func (r *Runner) Run(ctx context.Context, sess session.Session, opts RunOptions, events chan<- event.Event) error {
	if r.Provider == nil {
		return fmt.Errorf("provider is nil")
	}
	if r.Tools == nil {
		return fmt.Errorf("tools is nil")
	}
	if r.Context == nil {
		return fmt.Errorf("context manager is nil")
	}
	if r.Safety == nil {
		r.Safety = &NoOpSafetyLayer{}
	}

	send := func(ev event.Event) {
		select {
		case events <- ev:
		case <-ctx.Done():
		}
	}

	send(event.AgentStartEvent{AgentID: sess.ID()})
	defer send(event.AgentEndEvent{AgentID: sess.ID()})

	for turn := 1; opts.MaxTurns == 0 || turn <= opts.MaxTurns; turn++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		send(event.TurnStartEvent{Turn: turn})

		ctxLLM, cancel := context.WithTimeout(ctx, opts.Timeout)
		resp, err := r.step(ctxLLM, sess)
		cancel()
		if err != nil {
			send(event.ErrorEvent{Err: err})
			return err
		}

		// Append assistant message.
		if err := sess.Append(ctx, llm.Message{
			Role:      RoleAssistant,
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		}); err != nil {
			return err
		}

		// Stream text content as events.
		if resp.Content != "" {
			send(event.TextDeltaEvent{Delta: resp.Content})
		}

		if len(resp.ToolCalls) == 0 {
			send(event.TurnEndEvent{Turn: turn})
			return nil
		}

		// Execute tool calls.
		for _, call := range resp.ToolCalls {
			send(event.ToolCallStartEvent{Call: call})

			tool, ok := r.Tools.Get(call.Name)
			if !ok {
				send(event.ToolResultEvent{
					CallID: call.ID,
					Name:   call.Name,
					Result: tools.ToolResult{
						Metadata: tools.ToolResultMetadata{
							Error:  fmt.Sprintf("unknown tool: %s", call.Name),
							Status: tools.ToolExecuteStatusError,
						},
					},
				})
				continue
			}

			env := &toolEnv{
				fs:      sess.FileSystem(),
				sandbox: sess.Sandbox(),
				logger:  r.Logger,
			}

			allowed, err := r.Safety.BeforeToolCall(ctx, call, env)
			if err != nil || !allowed {
				errMsg := "blocked by safety layer"
				if err != nil {
					errMsg = err.Error()
				}
				send(event.ToolResultEvent{
					CallID: call.ID,
					Name:   call.Name,
					Result: tools.ToolResult{
						Metadata: tools.ToolResultMetadata{
							Error:  errMsg,
							Status: tools.ToolExecuteStatusError,
						},
					},
				})
				continue
			}

			res, err := tool.Execute(ctx, env, call.Arguments)
			if err != nil {
				res = tools.ToolResult{
					Metadata: tools.ToolResultMetadata{
						Error:  err.Error(),
						Status: tools.ToolExecuteStatusError,
					},
				}
			}

			res, _ = r.Safety.AfterToolCall(ctx, call, res)

			send(event.ToolResultEvent{
				CallID: call.ID,
				Name:   call.Name,
				Result: res,
			})

			// Append tool result to session.
			content := ""
			for _, part := range res.Content {
				content += part.Text
			}
			if res.Metadata.Error != "" {
				content = "error: " + res.Metadata.Error
			}
			_ = sess.Append(ctx, llm.Message{
				Role:       RoleTool,
				Content:    content,
				ToolCallID: call.ID,
				Name:       call.Name,
			})

			send(event.ToolCallEndEvent{CallID: call.ID})
		}

		send(event.TurnEndEvent{Turn: turn})
	}

	return fmt.Errorf("max turns exceeded")
}

func (r *Runner) step(ctx context.Context, sess session.Session) (*llm.Response, error) {
	ctxLLM, err := r.Context.Build(sess, r.Tools)
	if err != nil {
		return nil, err
	}

	resp, err := r.Provider.Complete(ctx, llm.Model{}, ctxLLM, llm.Options{})
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

type toolEnv struct {
	fs      fs.FileSystem
	sandbox sandbox.Sandbox
	logger  *slog.Logger
}

func (e *toolEnv) FileSystem() fs.FileSystem { return e.fs }
func (e *toolEnv) Sandbox() sandbox.Sandbox  { return e.sandbox }
func (e *toolEnv) Logger() *slog.Logger      { return e.logger }
