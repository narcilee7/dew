package plan

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/narcilee7/dew/pkg/ai"
	"github.com/narcilee7/dew/pkg/core"
	"github.com/narcilee7/dew/pkg/fs"
	"github.com/narcilee7/dew/pkg/sandbox"
	"github.com/narcilee7/dew/pkg/session"
	"github.com/narcilee7/dew/pkg/tools"
)

func TestPlanPluginCreatesCheckpoint(t *testing.T) {
	root := t.TempDir()
	fsys := fs.NewLocal(root)
	box := sandbox.NewLocal("test", fsys)
	store := session.NewMemoryStore()

	sess, _ := store.Create(context.Background(), session.CreateOptions{
		ID:      "session-plan",
		FS:      fsys,
		Sandbox: box,
	})

	// Register a mock bash tool so the loop actually executes it.
	registry := core.NewToolRegistry()
	_ = registry.Register(&mockTool{name: "bash"})

	callCount := 0
	provider := ai.NewMockProviderFunc(func(ctx context.Context, model ai.Model, context ai.Context, opts ai.ChatOptions) (*ai.Response, error) {
		callCount++
		if callCount > 1 {
			return &ai.Response{Content: "done"}, nil
		}
		return &ai.Response{
			Content: "running bash",
			ToolCalls: []ai.ToolCall{
				ai.MockToolCall("call-1", "bash", map[string]any{"command": "echo hi"}),
			},
		}, nil
	})

	planStore := NewFileStore(fsys)
	plugin := NewPluginWithStore(planStore)
	harness := core.NewHarness("test", core.Boundaries{
		Provider: provider,
		Tools:    registry,
	})
	_ = harness.Use(plugin)

	_ = sess.Append(context.Background(), ai.Message{Role: core.RoleUser, Content: "run bash"})

	done := make(chan error, 1)
	go func() {
		done <- harness.Run(context.Background(), sess, core.RunOptions{MaxTurns: 2})
	}()
	for range harness.Events() {
	}
	if err := <-done; err != nil {
		t.Fatalf("run: %v", err)
	}

	planID := "session-session-plan"
	cps, err := planStore.ListCheckpoints(context.Background(), planID)
	if err != nil {
		t.Fatalf("list checkpoints: %v", err)
	}
	if len(cps) != 1 {
		t.Fatalf("checkpoints = %v", cps)
	}
	if cps[0].Label != "before bash" {
		t.Fatalf("label = %q", cps[0].Label)
	}

	cp, err := planStore.LoadCheckpoint(context.Background(), planID, cps[0].ID)
	if err != nil {
		t.Fatalf("load checkpoint: %v", err)
	}
	if len(cp.Snapshot.Messages) == 0 {
		t.Fatal("expected snapshot messages")
	}
	if len(cp.Snapshot.Files) == 0 {
		t.Fatal("expected snapshot files")
	}
}

type mockTool struct {
	name string
}

func (m *mockTool) Name() string               { return m.name }
func (m *mockTool) Description() string        { return "mock tool" }
func (m *mockTool) Schema() tools.ToolSchema   { return tools.ToolSchema{} }
func (m *mockTool) Execute(ctx context.Context, env tools.ToolEnvironment, args json.RawMessage) (tools.ToolResult, error) {
	return tools.ToolResult{
		Content: []tools.ContentPart{{Type: "text", Text: "hi"}},
		Metadata: tools.ToolResultMetadata{Status: tools.ToolExecuteStatusSuccess},
	}, nil
}
