package soul

import (
	"context"

	"github.com/narcilee7/dew/pkg/core"
	"github.com/narcilee7/dew/pkg/fs"
	"github.com/narcilee7/dew/pkg/session"
)

// Plugin is a core.Plugin that loads the Soul before a session and observes
// events for candidate updates.
type Plugin struct {
	engine Engine
}

// NewPlugin creates a soul plugin backed by the given filesystem.
func NewPlugin(fsys fs.FileSystem, path string) *Plugin {
	store := NewFileStore(fsys, path)
	return &Plugin{engine: NewSimpleEngine(store)}
}

// NewPluginWithEngine creates a soul plugin with a custom engine.
func NewPluginWithEngine(engine Engine) *Plugin {
	return &Plugin{engine: engine}
}

// Name returns the plugin name.
func (p *Plugin) Name() string { return "soul" }

// Hooks returns the hooks provided by the plugin.
func (p *Plugin) Hooks() []core.Hook {
	return []core.Hook{
		&soulHook{p: p, point: core.HookBeforeSession},
		&soulHook{p: p, point: core.HookAfterSession},
		&soulHook{p: p, point: core.HookOnEvent},
	}
}

type soulHook struct {
	p     *Plugin
	point core.HookPoint
}

func (h *soulHook) Point() core.HookPoint { return h.point }

func (h *soulHook) Handle(ctx context.Context, env core.HookEnvironment) error {
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
	soul, err := p.engine.Load(ctx)
	if err != nil {
		return err
	}
	task := firstUserMessage(env.Session)
	fragment, err := p.engine.Mix(ctx, soul, task)
	if err != nil {
		return err
	}
	return core.InjectSystemPromptFragment(env.Harness, fragment)
}

func (p *Plugin) afterSession(ctx context.Context, env core.HookEnvironment) error {
	// Future: flush buffered updates, reflect, and apply approved ones.
	return nil
}

func (p *Plugin) onEvent(ctx context.Context, env core.HookEnvironment) error {
	// Future: observe events and buffer candidate Soul updates.
	return nil
}

func firstUserMessage(sess session.Session) string {
	for _, m := range sess.Messages() {
		if m.Role == "user" {
			return m.Content
		}
	}
	return ""
}
