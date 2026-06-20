package memory

import (
	"context"
	"testing"

	"github.com/narcilee7/dew/pkg/ai"
	"github.com/narcilee7/dew/pkg/core"
	"github.com/narcilee7/dew/pkg/fs"
	"github.com/narcilee7/dew/pkg/sandbox"
	"github.com/narcilee7/dew/pkg/session"
)

func TestMemoryPluginInjectsRelevantMemories(t *testing.T) {
	root := t.TempDir()
	fsys := fs.NewLocal(root)
	box := sandbox.NewLocal("test", fsys)
	store := session.NewMemoryStore()

	sess, _ := store.Create(context.Background(), session.CreateOptions{
		ID:      "session-mem",
		FS:      fsys,
		Sandbox: box,
	})

	// Seed a memory.
	memStore := NewFileStore(fsys)
	_ = memStore.Remember(context.Background(), Memory{
		ID:      "m1",
		Type:    MemoryTypeFact,
		Content: "The project uses Go interfaces for boundaries.",
	})

	plugin := NewPluginWithStore(memStore)
	harness := core.NewHarness("test", core.Boundaries{
		Provider: ai.NewMockProviderFunc(func(ctx context.Context, model ai.Model, context ai.Context, opts ai.ChatOptions) (*ai.Response, error) {
			return &ai.Response{Content: "done"}, nil
		}),
		Tools: core.NewToolRegistry(),
	})
	_ = harness.Use(plugin)

	_ = sess.Append(context.Background(), ai.Message{Role: core.RoleUser, Content: "tell me about boundaries"})

	done := make(chan error, 1)
	go func() {
		done <- harness.Run(context.Background(), sess, core.RunOptions{MaxTurns: 1})
	}()
	for range harness.Events() {
	}
	if err := <-done; err != nil {
		t.Fatalf("run: %v", err)
	}

	loop := harness.Loop().(*core.DefaultLoop)
	fragments := loop.SystemPromptFragments()
	if len(fragments) != 1 {
		t.Fatalf("fragments = %v", fragments)
	}
	if fragments[0] == "" {
		t.Fatal("expected non-empty memory fragment")
	}
}
