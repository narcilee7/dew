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

// Loop executes the agent turn sequence for a Harness.
// Custom loops can replace DefaultLoop to implement different agent behaviors.
type Loop interface {
	Run(ctx context.Context, h Harness, sess session.Session, opts RunOptions) error
}

// DefaultLoop is the standard ReAct loop.
type DefaultLoop struct {
	Safety  SafetyLayer
	Context ContextManager
	Logger  *slog.Logger
}

// Run executes the default agent loop.
func (l *DefaultLoop) Run(ctx context.Context, h Harness, sess session.Session, opts RunOptions) error {
	b := h.Boundaries()
	if b.Provider == nil {
		return fmt.Errorf("provider is nil")
	}
	if b.Tools == nil {
		return fmt.Errorf("tools is nil")
	}
	if l.Context == nil {
		return fmt.Errorf("context manager is nil")
	}
	if l.Safety == nil {
		l.Safety = &NoOpSafetyLayer{}
	}
	if l.Logger == nil {
		l.Logger = slog.Default()
	}

	h.Emit(event.AgentStartEvent{AgentID: sess.ID()})
	defer h.Emit(event.AgentEndEvent{AgentID: sess.ID()})

	for turn := 1; opts.MaxTurns == 0 || turn <= opts.MaxTurns; turn++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := h.RunHooks(ctx, HookBeforeTurn, HookEnvironment{Harness: h, Session: sess, Turn: turn}); err != nil {
			return err
		}

		h.Emit(event.TurnStartEvent{Turn: turn})

		resp, err := l.step(ctx, h, sess)
		if err != nil {
			h.Emit(event.ErrorEvent{Err: err})
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
			h.Emit(event.TextDeltaEvent{Delta: resp.Content})
		}

		if len(resp.ToolCalls) == 0 {
			h.Emit(event.TurnEndEvent{Turn: turn})
			if err := h.RunHooks(ctx, HookAfterTurn, HookEnvironment{Harness: h, Session: sess, Turn: turn}); err != nil {
				return err
			}
			return nil
		}

		// Execute tool calls.
		for _, call := range resp.ToolCalls {
			h.Emit(event.ToolCallStartEvent{Call: call})

			tool, ok := b.Tools.Get(call.Name)
			if !ok {
				h.Emit(event.ToolResultEvent{
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
				logger:  l.Logger,
			}

			if err := h.RunHooks(ctx, HookBeforeToolUse, HookEnvironment{Harness: h, Session: sess, Turn: turn, Call: &call}); err != nil {
				h.Emit(event.ToolResultEvent{
					CallID: call.ID,
					Name:   call.Name,
					Result: tools.ToolResult{
						Metadata: tools.ToolResultMetadata{
							Error:  err.Error(),
							Status: tools.ToolExecuteStatusError,
						},
					},
				})
				continue
			}

			allowed, err := l.Safety.BeforeToolCall(ctx, call, env)
			if err != nil || !allowed {
				errMsg := "blocked by safety layer"
				if err != nil {
					errMsg = err.Error()
				}
				h.Emit(event.ToolResultEvent{
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

			res, _ = l.Safety.AfterToolCall(ctx, call, res)

			if err := h.RunHooks(ctx, HookAfterToolUse, HookEnvironment{Harness: h, Session: sess, Turn: turn, Call: &call, Result: &res}); err != nil {
				h.Emit(event.ErrorEvent{Err: err})
				return err
			}

			h.Emit(event.ToolResultEvent{
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

			h.Emit(event.ToolCallEndEvent{CallID: call.ID})
		}

		h.Emit(event.TurnEndEvent{Turn: turn})
		if err := h.RunHooks(ctx, HookAfterTurn, HookEnvironment{Harness: h, Session: sess, Turn: turn}); err != nil {
			return err
		}
	}

	return fmt.Errorf("max turns exceeded")
}

func (l *DefaultLoop) step(ctx context.Context, h Harness, sess session.Session) (*llm.Response, error) {
	b := h.Boundaries()
	ctxLLM, err := l.Context.Build(sess, b.Tools)
	if err != nil {
		return nil, err
	}

	resp, err := b.Provider.Complete(ctx, llm.Model{}, ctxLLM, llm.Options{})
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
