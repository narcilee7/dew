package skill

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/narcilee7/dew/pkg/core"
)

// promptHook injects a prompt fragment into the default loop at a hook point.
type promptHook struct {
	point core.HookPoint
	skill *Skill
	file  string
}

func (h *promptHook) Point() core.HookPoint {
	return h.point
}

func (h *promptHook) Handle(ctx context.Context, env core.HookEnvironment) error {
	dn, ok := env.Harness.(*core.DefaultHarness)
	if !ok {
		return nil
	}
	loop, ok := dn.Loop().(*core.DefaultLoop)
	if !ok {
		return nil
	}
	path := filepath.Join(h.skill.Path, h.file)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read skill prompt %q: %w", h.file, err)
	}
	if len(data) == 0 {
		return nil
	}

	loop.AddSystemPromptFragment(string(data))

	return nil
}

func parseHookPoint(s string) core.HookPoint {
	switch s {
	case "before_session":
		return core.HookBeforeSession
	case "after_session":
		return core.HookAfterSession
	case "before_turn":
		return core.HookBeforeTurn
	case "after_turn":
		return core.HookAfterTurn
	case "before_tool_use":
		return core.HookBeforeToolUse
	case "after_tool_use":
		return core.HookAfterToolUse
	case "on_event":
		return core.HookOnEvent
	case "on_error":
		return core.HookOnError
	default:
		return -1
	}
}
