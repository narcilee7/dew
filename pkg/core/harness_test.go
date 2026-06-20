package core

import (
	"context"
	"testing"

	"github.com/narcilee7/dew/pkg/fs"
	"github.com/narcilee7/dew/pkg/llm"
	"github.com/narcilee7/dew/pkg/sandbox"
	"github.com/narcilee7/dew/pkg/session"
	"github.com/narcilee7/dew/pkg/tools"
)

// recordingPlugin records every hook point it sees.
type recordingPlugin struct {
	points []HookPoint
}

func (p *recordingPlugin) Name() string { return "recording" }

func (p *recordingPlugin) Hooks() []Hook {
	return []Hook{
		&recordHook{p: p, point: HookBeforeSession},
		&recordHook{p: p, point: HookAfterSession},
		&recordHook{p: p, point: HookBeforeTurn},
		&recordHook{p: p, point: HookAfterTurn},
		&recordHook{p: p, point: HookBeforeToolUse},
		&recordHook{p: p, point: HookAfterToolUse},
	}
}

type recordHook struct {
	p     *recordingPlugin
	point HookPoint
}

func (h *recordHook) Point() HookPoint { return h.point }

func (h *recordHook) Handle(ctx context.Context, env HookEnvironment) error {
	h.p.points = append(h.p.points, h.point)
	return nil
}

func TestHarnessRunsPluginsAndLoop(t *testing.T) {
	root := t.TempDir()
	fsys := fs.NewLocal(root)
	box := sandbox.NewLocal("test", fsys)
	store := session.NewMemoryStore()

	sess, err := store.Create(context.Background(), session.CreateOptions{
		ID:      "session-test",
		FS:      fsys,
		Sandbox: box,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	registry := NewToolRegistry()
	_ = registry.Register(&tools.BashTool{})

	provider := llm.NewMockProviderFunc(func(ctx context.Context, model llm.Model, context llm.Context, opts llm.Options) (*llm.Response, error) {
		return &llm.Response{
			Content: "hello",
		}, nil
	})

	rec := &recordingPlugin{}
	harness := NewHarness("test", Boundaries{
		Provider: provider,
		Tools:    registry,
		Session:  store,
	})
	if err := harness.Use(rec); err != nil {
		t.Fatalf("use plugin: %v", err)
	}

	_ = sess.Append(context.Background(), llm.Message{Role: RoleUser, Content: "hi"})

	done := make(chan error, 1)
	go func() {
		done <- harness.Run(context.Background(), sess, RunOptions{MaxTurns: 5})
	}()

	events := 0
	for range harness.Events() {
		events++
	}

	if err := <-done; err != nil {
		t.Fatalf("run: %v", err)
	}

	if events == 0 {
		t.Fatalf("expected events, got none")
	}

	want := []HookPoint{HookBeforeSession, HookBeforeTurn, HookAfterTurn, HookAfterSession}
	if len(rec.points) != len(want) {
		t.Fatalf("hook points = %v, want %v", rec.points, want)
	}
	for i, p := range want {
		if rec.points[i] != p {
			t.Fatalf("hook point %d = %v, want %v", i, rec.points[i], p)
		}
	}
}

func TestHarnessToolHooksFire(t *testing.T) {
	root := t.TempDir()
	fsys := fs.NewLocal(root)
	box := sandbox.NewLocal("test", fsys)
	store := session.NewMemoryStore()

	sess, err := store.Create(context.Background(), session.CreateOptions{
		ID:      "session-tool-test",
		FS:      fsys,
		Sandbox: box,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	registry := NewToolRegistry()
	_ = registry.Register(&tools.BashTool{})

	provider := llm.NewMockProviderFunc(func(ctx context.Context, model llm.Model, context llm.Context, opts llm.Options) (*llm.Response, error) {
		for i := len(context.Messages) - 1; i >= 0; i-- {
			m := context.Messages[i]
			if m.Role == RoleAssistant && len(m.ToolCalls) > 0 {
				return &llm.Response{Content: "done"}, nil
			}
			if m.Role == RoleUser {
				break
			}
		}
		return &llm.Response{
			Content: "running tool",
			ToolCalls: []llm.ToolCall{
				llm.MockToolCall("call-1", "bash", map[string]any{"command": "echo ok"}),
			},
		}, nil
	})

	rec := &recordingPlugin{}
	harness := NewHarness("test", Boundaries{
		Provider: provider,
		Tools:    registry,
		Session:  store,
	})
	_ = harness.Use(rec)

	_ = sess.Append(context.Background(), llm.Message{Role: RoleUser, Content: "run"})

	done := make(chan error, 1)
	go func() {
		done <- harness.Run(context.Background(), sess, RunOptions{MaxTurns: 5})
	}()

	for range harness.Events() {
	}

	if err := <-done; err != nil {
		t.Fatalf("run: %v", err)
	}

	hasBeforeToolUse := false
	hasAfterToolUse := false
	for _, p := range rec.points {
		if p == HookBeforeToolUse {
			hasBeforeToolUse = true
		}
		if p == HookAfterToolUse {
			hasAfterToolUse = true
		}
	}
	if !hasBeforeToolUse {
		t.Fatalf("expected HookBeforeToolUse to fire")
	}
	if !hasAfterToolUse {
		t.Fatalf("expected HookAfterToolUse to fire")
	}
}

