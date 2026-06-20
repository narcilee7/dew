package trajectory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/narcilee7/dew/pkg/core"
	"github.com/narcilee7/dew/pkg/fs"
)

// Plugin records harness events into a persistent trajectory.
type Plugin struct {
	store Store

	mu    sync.Mutex
	traj  *Trajectory
	seq   int
}

// NewPlugin creates a trajectory plugin backed by the given filesystem.
func NewPlugin(fsys fs.FileSystem) *Plugin {
	return &Plugin{store: NewFileStore(fsys)}
}

// NewPluginWithStore creates a trajectory plugin with a custom store.
func NewPluginWithStore(store Store) *Plugin {
	return &Plugin{store: store}
}

// Name returns the plugin name.
func (p *Plugin) Name() string { return "trajectory" }

// Hooks returns the hooks provided by the plugin.
func (p *Plugin) Hooks() []core.Hook {
	return []core.Hook{
		&trajectoryHook{p: p, point: core.HookBeforeSession},
		&trajectoryHook{p: p, point: core.HookAfterSession},
		&trajectoryHook{p: p, point: core.HookOnEvent},
	}
}

type trajectoryHook struct {
	p     *Plugin
	point core.HookPoint
}

func (h *trajectoryHook) Point() core.HookPoint { return h.point }

func (h *trajectoryHook) Handle(ctx context.Context, env core.HookEnvironment) error {
	switch h.point {
	case core.HookBeforeSession:
		return h.p.beforeSession(ctx, env)
	case core.HookAfterSession:
		return h.p.afterSession(ctx, env)
	case core.HookOnEvent:
		return h.p.onEvent(ctx, env)
	}
	return nil
}

func (p *Plugin) beforeSession(ctx context.Context, env core.HookEnvironment) error {
	now := time.Now()
	traj := Trajectory{
		ID:        fmt.Sprintf("traj-%s", env.Session.ID()),
		SessionID: env.Session.ID(),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := p.store.Create(ctx, traj); err != nil {
		return fmt.Errorf("create trajectory: %w", err)
	}

	p.mu.Lock()
	p.traj = &traj
	p.seq = 0
	p.mu.Unlock()
	return nil
}

func (p *Plugin) afterSession(ctx context.Context, env core.HookEnvironment) error {
	p.mu.Lock()
	traj := p.traj
	p.mu.Unlock()
	if traj == nil {
		return nil
	}
	traj.UpdatedAt = time.Now()
	return p.store.Save(ctx, *traj)
}

func (p *Plugin) onEvent(ctx context.Context, env core.HookEnvironment) error {
	if env.Event == nil {
		return nil
	}

	p.mu.Lock()
	traj := p.traj
	if traj == nil {
		p.mu.Unlock()
		return nil
	}
	p.seq++
	seq := p.seq
	p.mu.Unlock()

	rec, err := NewEventRecord(env.Event, seq, env.Session.ID(), env.Turn)
	if err != nil {
		return fmt.Errorf("encode event: %w", err)
	}
	return p.store.AppendEvent(ctx, traj.ID, rec)
}
