package memory

import (
	"context"
	"fmt"
	"strings"

	"github.com/narcilee7/dew/pkg/core"
	"github.com/narcilee7/dew/pkg/fs"
	"github.com/narcilee7/dew/pkg/session"
)

// Plugin is a core.Plugin that loads relevant memories before a session and
// persists learned memories after a session.
type Plugin struct {
	store Store
}

// NewPlugin creates a memory plugin backed by the given filesystem.
func NewPlugin(fsys fs.FileSystem) *Plugin {
	return &Plugin{store: NewFileStore(fsys)}
}

// NewPluginWithStore creates a memory plugin with a custom store.
func NewPluginWithStore(store Store) *Plugin {
	return &Plugin{store: store}
}

// Name returns the plugin name.
func (p *Plugin) Name() string { return "memory" }

// Hooks returns the hooks provided by the plugin.
func (p *Plugin) Hooks() []core.Hook {
	return []core.Hook{
		&memoryHook{p: p, point: core.HookBeforeSession},
		&memoryHook{p: p, point: core.HookAfterSession},
	}
}

type memoryHook struct {
	p     *Plugin
	point core.HookPoint
}

func (h *memoryHook) Point() core.HookPoint { return h.point }

func (h *memoryHook) Handle(ctx context.Context, env core.HookEnvironment) error {
	switch h.point {
	case core.HookBeforeSession:
		return h.p.beforeSession(ctx, env)
	case core.HookAfterSession:
		return h.p.afterSession(ctx, env)
	}
	return nil
}

func (p *Plugin) beforeSession(ctx context.Context, env core.HookEnvironment) error {
	query := firstUserMessage(env.Session)
	if query == "" {
		return nil
	}

	memories, err := p.store.Recall(ctx, RecallQuery{
		Query: query,
		Types: []MemoryType{MemoryTypeRule, MemoryTypeFact, MemoryTypePreference, MemoryTypeLesson},
		Limit: 5,
	})
	if err != nil {
		return fmt.Errorf("recall memories: %w", err)
	}
	if len(memories) == 0 {
		return nil
	}

	var b strings.Builder
	b.WriteString("## Relevant memories\n\n")
	for _, mem := range memories {
		b.WriteString(fmt.Sprintf("- [%s] %s\n", mem.Type, mem.Content))
	}
	return core.InjectSystemPromptFragment(env.Harness, b.String())
}

func (p *Plugin) afterSession(ctx context.Context, env core.HookEnvironment) error {
	// Future: distill session transcript into memories and call Remember.
	// For now, the plugin does not auto-write memories to avoid noise.
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
