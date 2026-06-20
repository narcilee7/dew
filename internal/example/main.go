package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/narcilee7/dew/pkg/core"
	"github.com/narcilee7/dew/pkg/event"
	"github.com/narcilee7/dew/pkg/fs"
	"github.com/narcilee7/dew/pkg/llm"
	"github.com/narcilee7/dew/pkg/sandbox"
	"github.com/narcilee7/dew/pkg/session"
	"github.com/narcilee7/dew/pkg/tools"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// 1. Create isolated filesystem.
	root, err := os.MkdirTemp("", "dew-*")
	if err != nil {
		logger.Error("create root", "error", err)
		os.Exit(1)
	}
	defer os.RemoveAll(root)

	fsys := fs.NewLocal(root)

	// 2. Create sandbox.
	box := sandbox.NewLocal("main", fsys)

	// 3. Create session.
	store := session.NewMemoryStore()
	sess, err := store.Create(context.Background(), session.CreateOptions{
		ID:      "session-1",
		FS:      fsys,
		Sandbox: box,
	})
	if err != nil {
		logger.Error("create session", "error", err)
		os.Exit(1)
	}

	// 4. Register tools.
	registry := core.NewToolRegistry()
	_ = registry.Register(&tools.ReadTool{})
	_ = registry.Register(&tools.BashTool{})

	// 5. Create a mock provider that triggers a bash tool call once, then stops.
	provider := llm.NewMockProviderFunc(func(ctx context.Context, model llm.Model, context llm.Context, opts llm.Options) (*llm.Response, error) {
		// If the last assistant message already had tool calls, finish.
		for i := len(context.Messages) - 1; i >= 0; i-- {
			m := context.Messages[i]
			if m.Role == core.RoleAssistant && len(m.ToolCalls) > 0 {
				return &llm.Response{Content: "Done."}, nil
			}
			if m.Role == core.RoleUser {
				break
			}
		}
		return &llm.Response{
			Content: "I'll run a command for you.",
			ToolCalls: []llm.ToolCall{
				llm.MockToolCall("call-1", "bash", map[string]any{"command": "echo hello from dew"}),
			},
		}, nil
	})

	// 6. Create runner.
	runner := core.NewRunner(provider, registry)
	runner.Logger = logger
	runner.Context = &core.SimpleContextManager{
		SystemPrompt: "You are dew, a helpful coding agent.",
	}

	// 7. Seed user message.
	_ = sess.Append(context.Background(), llm.Message{
		Role:    core.RoleUser,
		Content: "Say hello",
	})

	// 8. Run and observe events.
	events := make(chan event.Event, 64)
	done := make(chan error, 1)

	go func() {
		done <- runner.Run(context.Background(), sess, core.RunOptions{
			MaxTurns: 10,
			Timeout:  30 * time.Second,
		}, events)
		close(events)
	}()

	fmt.Println("=== dew events ===")
	for ev := range events {
		switch e := ev.(type) {
		case event.AgentStartEvent:
			fmt.Printf("[start] agent=%s\n", e.AgentID)
		case event.TurnStartEvent:
			fmt.Printf("[turn] %d\n", e.Turn)
		case event.TextDeltaEvent:
			fmt.Printf("[text] %s\n", e.Delta)
		case event.ToolCallStartEvent:
			fmt.Printf("[tool-start] %s(%s)\n", e.Call.Name, e.Call.ID)
		case event.ToolResultEvent:
			for _, part := range e.Result.Content {
				fmt.Printf("[tool-result] %s: %s\n", e.Name, part.Text)
			}
			if e.Result.Metadata.Error != "" {
				fmt.Printf("[tool-error] %s: %s\n", e.Name, e.Result.Metadata.Error)
			}
		case event.TurnEndEvent:
			fmt.Printf("[/turn] %d\n", e.Turn)
		case event.AgentEndEvent:
			fmt.Printf("[end] agent=%s\n", e.AgentID)
		case event.ErrorEvent:
			fmt.Printf("[error] %v\n", e.Err)
		}
	}

	if err := <-done; err != nil {
		logger.Error("run failed", "error", err)
		os.Exit(1)
	}
}
