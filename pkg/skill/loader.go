package skill

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Loader discovers and loads skills from disk.
type Loader struct {
	// ProjectPaths are directories that may contain .dew/skills/<name>/ subdirs.
	ProjectPaths []string

	// UserPaths are directories that may contain skills/<name>/ subdirs.
	UserPaths []string
}

// NewLoader creates a loader with default paths.
func NewLoader() *Loader {
	home, _ := os.UserHomeDir()
	return &Loader{
		ProjectPaths: []string{"."},
		UserPaths:    []string{filepath.Join(home, ".config", "dew", "skills")},
	}
}

// Load discovers and loads all skills.
func (l *Loader) Load(ctx context.Context) ([]Skill, error) {
	var skills []Skill
	seen := make(map[string]bool)

	loadDir := func(base string, skillSubdir string) error {
		dir := filepath.Join(base, skillSubdir)
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			skillDir := filepath.Join(dir, entry.Name())
			s, err := LoadSkill(skillDir)
			if err != nil {
				return fmt.Errorf("load skill %q: %w", entry.Name(), err)
			}
			if seen[s.Manifest.Name] {
				continue
			}
			seen[s.Manifest.Name] = true
			skills = append(skills, *s)
		}
		return nil
	}

	for _, base := range l.ProjectPaths {
		if err := loadDir(base, ".dew/skills"); err != nil {
			return nil, err
		}
	}
	for _, base := range l.UserPaths {
		if err := loadDir(base, ""); err != nil {
			return nil, err
		}
	}

	return skills, nil
}

// LoadSkill loads a single skill directory.
func LoadSkill(path string) (*Skill, error) {
	manifestPath := filepath.Join(path, "manifest.toml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	var manifest Manifest
	if err := toml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	if manifest.Name == "" {
		manifest.Name = filepath.Base(path)
	}
	if manifest.Version == "" {
		manifest.Version = "0.0.0"
	}

	docPath := filepath.Join(path, "SKILL.md")
	doc, err := os.ReadFile(docPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("read SKILL.md: %w", err)
	}

	return &Skill{
		Path:     path,
		Manifest: manifest,
		Document: string(doc),
	}, nil
}

