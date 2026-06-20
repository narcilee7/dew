package skill

import "github.com/narcilee7/dew/pkg/core"

// Manifest is the machine-readable description of a skill.
type Manifest struct {
	Name           string            `toml:"name"`
	Version        string            `toml:"version"`
	Description    string            `toml:"description"`
	Author         string            `toml:"author,omitempty"`
	SystemPrompt   string            `toml:"system_prompt,omitempty"`
	Hooks          map[string]string `toml:"hooks,omitempty"`
	Metadata       map[string]string `toml:"metadata,omitempty"`
}

// Skill represents a loaded skill package.
type Skill struct {
	// Path is the directory containing the skill.
	Path string

	// Manifest is the parsed manifest.toml.
	Manifest Manifest

	// Document is the raw content of SKILL.md.
	Document string
}

// Plugin converts the skill into a core.Plugin.
func (s *Skill) Plugin() core.Plugin {
	return &skillPlugin{skill: s}
}

// Name returns the skill name.
func (s *Skill) Name() string { return s.Manifest.Name }

// Version returns the skill version.
func (s *Skill) Version() string { return s.Manifest.Version }
