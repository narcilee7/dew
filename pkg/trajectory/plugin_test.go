package trajectory

import (
	"context"
	"testing"

	"github.com/narcilee7/dew/pkg/ai"
	"github.com/narcilee7/dew/pkg/core"
	"github.com/narcilee7/dew/pkg/event"
	"github.com/narcilee7/dew/pkg/fs"
	"github.com/narcilee7/dew/pkg/sandbox"
	"github.com/narcilee7/dew/pkg/session"
)

func TestTrajectoryPluginRecordsEvents(t *testing.T) {
	root := t.TempDir()
	fsys := fs.NewLocal(root)
	box := sandbox.NewLocal("test", fsys)
	store := session.NewMemoryStore()

	sess, _ := store.Create(context.Background(), session.CreateOptions{
		ID:      "session-traj",
		FS:      fsys,
		Sandbox: box,
	})

	tjStore := NewFileStore(fsys)
	plugin := NewPluginWithStore(tjStore)
	harness := core.NewHarness("test", core.Boundaries{
		Provider: ai.NewMockProviderFunc(func(ctx context.Context, model ai.Model, context ai.Context, opts ai.ChatOptions) (*ai.Response, error) {
			return &ai.Response{Content: "done"}, nil
		}),
		Tools: core.NewToolRegistry(),
	})
	_ = harness.Use(plugin)

	_ = sess.Append(context.Background(), ai.Message{Role: core.RoleUser, Content: "hello"})

	done := make(chan error, 1)
	go func() {
		done <- harness.Run(context.Background(), sess, core.RunOptions{MaxTurns: 1})
	}()
	for range harness.Events() {
	}
	if err := <-done; err != nil {
		t.Fatalf("run: %v", err)
	}

	traj, err := tjStore.Load(context.Background(), "traj-session-traj")
	if err != nil {
		t.Fatalf("load trajectory: %v", err)
	}
	if traj.SessionID != "session-traj" {
		t.Fatalf("session id = %q", traj.SessionID)
	}

	events, err := tjStore.ListEvents(context.Background(), traj.ID)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) == 0 {
		t.Fatal("expected events")
	}

	var sawAgentStart, sawAgentEnd bool
	for _, rec := range events {
		ev, err := rec.DecodePayload()
		if err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		switch ev.(type) {
		case *event.AgentStartEvent:
			sawAgentStart = true
		case *event.AgentEndEvent:
			sawAgentEnd = true
		}
	}
	if !sawAgentStart || !sawAgentEnd {
		t.Fatalf("missing lifecycle events: start=%v end=%v", sawAgentStart, sawAgentEnd)
	}
}
