package skill

import (
	"github.com/narcilee7/dew/pkg/core"
)

// skillPlugin adapts a Skill to the core.Plugin interface.
type skillPlugin struct {
	skill *Skill
}

// Name returns the plugin name.
func (p *skillPlugin) Name() string {
	return "skill:" + p.skill.Manifest.Name
}

// Hooks returns the hooks declared by the skill.
func (p *skillPlugin) Hooks() []core.Hook {
	var hooks []core.Hook

	// system_prompt is shorthand for before_session.
	if p.skill.Manifest.SystemPrompt != "" {
		hooks = append(hooks, &promptHook{
			point: core.HookBeforeSession,
			skill: p.skill,
			file:  p.skill.Manifest.SystemPrompt,
		})
	}

	for pointName, file := range p.skill.Manifest.Hooks {
		point := parseHookPoint(pointName)
		if point < 0 {
			continue
		}
		hooks = append(hooks, &promptHook{
			point: point,
			skill: p.skill,
			file:  file,
		})
	}

	return hooks
}
