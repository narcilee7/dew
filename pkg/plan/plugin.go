package plan

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/narcilee7/dew/pkg/core"
	"github.com/narcilee7/dew/pkg/fs"
	"github.com/narcilee7/dew/pkg/session"
)

// riskyTools are tools that mutate state and benefit from a checkpoint.
var riskyTools = map[string]bool{
	"write": true,
	"edit":  true,
	"bash":  true,
}

// Plugin is a core.Plugin that creates checkpoints before risky operations.
type Plugin struct {
	store Store
}

// NewPlugin creates a plan plugin backed by the given filesystem.
func NewPlugin(fsys fs.FileSystem) *Plugin {
	return &Plugin{store: NewFileStore(fsys)}
}

// NewPluginWithStore creates a plan plugin with a custom store.
func NewPluginWithStore(store Store) *Plugin {
	return &Plugin{store: store}
}

// Name returns the plugin name.
func (p *Plugin) Name() string { return "plan" }

// Hooks returns the hooks provided by the plugin.
func (p *Plugin) Hooks() []core.Hook {
	return []core.Hook{
		&planHook{p: p, point: core.HookBeforeToolUse},
	}
}

type planHook struct {
	p     *Plugin
	point core.HookPoint
}

func (h *planHook) Point() core.HookPoint { return h.point }

func (h *planHook) Handle(ctx context.Context, env core.HookEnvironment) error {
	if env.Call == nil {
		return nil
	}
	if !riskyTools[env.Call.Name] {
		return nil
	}

	// Create a lightweight checkpoint for the current session.
	planID := fmt.Sprintf("session-%s", env.Session.ID())
	cp := Checkpoint{
		ID:        fmt.Sprintf("cp-%d", time.Now().UnixNano()),
		PlanID:    planID,
		StepID:    env.Call.ID,
		Label:     fmt.Sprintf("before %s", env.Call.Name),
		CreatedAt: time.Now(),
	}

	// Ensure a plan exists for this session, then save the checkpoint.
	_, err := h.p.store.Load(ctx, planID)
	if err != nil {
		_ = h.p.store.Create(ctx, Plan{
			ID:        planID,
			Goal:      "session checkpoints",
			SessionID: env.Session.ID(),
			Steps:     nil,
			CreatedAt: time.Now(),
		})
	}

	cp.Snapshot = captureSnapshot(ctx, env.Session)
	return h.p.store.SaveCheckpoint(ctx, planID, cp)
}

// captureSnapshot records a lightweight snapshot of session state and file hashes.
func captureSnapshot(ctx context.Context, sess session.Session) Snapshot {
	messages, _ := json.Marshal(sess.Messages())
	files := make(map[string]string)
	_ = hashFiles(ctx, sess.FileSystem(), ".", files)
	return Snapshot{
		Messages:  messages,
		Files:     files,
		EventIndex: 0,
	}
}

func hashFiles(ctx context.Context, fsys fs.FileSystem, dir string, out map[string]string) error {
	entries, err := fsys.List(ctx, dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		path := e.Path
		if e.IsDir {
			_ = hashFiles(ctx, fsys, path, out)
			continue
		}
		data, err := fsys.Read(ctx, path)
		if err != nil {
			continue
		}
		sum := sha256.Sum256(data)
		out[path] = hex.EncodeToString(sum[:])
	}
	return nil
}
