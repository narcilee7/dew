package skill

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/narcilee7/dew/pkg/core"
	"github.com/narcilee7/dew/pkg/fs"
	"github.com/narcilee7/dew/pkg/ai"
	"github.com/narcilee7/dew/pkg/sandbox"
	"github.com/narcilee7/dew/pkg/session"
)

func TestLoadSkill(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "test-skill")
	_ = os.MkdirAll(filepath.Join(skillDir, "prompts"), 0755)

	manifest := `
name = "test-skill"
version = "0.1.0"
description = "A test skill"
system_prompt = "prompts/system.md"

[hooks]
before_turn = "prompts/before-turn.md"
`
	if err := os.WriteFile(filepath.Join(skillDir, "manifest.toml"), []byte(manifest), 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Test Skill\n"), 0644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "prompts", "system.md"), []byte("System prompt fragment."), 0644); err != nil {
		t.Fatalf("write system.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "prompts", "before-turn.md"), []byte("Before turn fragment."), 0644); err != nil {
		t.Fatalf("write before-turn.md: %v", err)
	}

	s, err := LoadSkill(skillDir)
	if err != nil {
		t.Fatalf("load skill: %v", err)
	}

	if s.Manifest.Name != "test-skill" {
		t.Fatalf("name = %q, want test-skill", s.Manifest.Name)
	}
	if s.Manifest.Version != "0.1.0" {
		t.Fatalf("version = %q, want 0.1.0", s.Manifest.Version)
	}
	if s.Manifest.SystemPrompt != "prompts/system.md" {
		t.Fatalf("system_prompt = %q", s.Manifest.SystemPrompt)
	}
	if len(s.Manifest.Hooks) != 1 {
		t.Fatalf("hooks = %v", s.Manifest.Hooks)
	}
	if s.Document != "# Test Skill\n" {
		t.Fatalf("document = %q", s.Document)
	}
}

func TestSkillPluginInjectsPrompt(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "refactor-go")
	_ = os.MkdirAll(filepath.Join(skillDir, "prompts"), 0755)

	manifest := `name = "refactor-go"
version = "0.1.0"
description = "Refactor Go code"
system_prompt = "prompts/system.md"
`
	_ = os.WriteFile(filepath.Join(skillDir, "manifest.toml"), []byte(manifest), 0644)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# refactor-go\n"), 0644)
	_ = os.WriteFile(filepath.Join(skillDir, "prompts", "system.md"), []byte("Always run go test."), 0644)

	s, err := LoadSkill(skillDir)
	if err != nil {
		t.Fatalf("load skill: %v", err)
	}

	root := t.TempDir()
	fsys := fs.NewLocal(root)
	box := sandbox.NewLocal("test", fsys)
	store := session.NewMemoryStore()
	sess, _ := store.Create(context.Background(), session.CreateOptions{
		ID:      "session-skill",
		FS:      fsys,
		Sandbox: box,
	})

	registry := core.NewToolRegistry()
	provider := ai.NewMockProviderFunc(func(ctx context.Context, model ai.Model, context ai.Context, opts ai.ChatOptions) (*ai.Response, error) {
		return &ai.Response{Content: "done"}, nil
	})

	harness := core.NewHarness("test", core.Boundaries{
		Provider: provider,
		Tools:    registry,
		Session:  store,
	})
	_ = harness.Use(s.Plugin())

	_ = sess.Append(context.Background(), ai.Message{Role: core.RoleUser, Content: "refactor"})

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
	if fragments[0] != "Always run go test." {
		t.Fatalf("fragment = %q", fragments[0])
	}
}

func TestLoaderDiscoversSkills(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, ".dew", "skills", "found")
	_ = os.MkdirAll(skillDir, 0755)
	manifest := `name = "found"
version = "1.0.0"
description = "Found skill"
`
	_ = os.WriteFile(filepath.Join(skillDir, "manifest.toml"), []byte(manifest), 0644)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Found\n"), 0644)

	loader := &Loader{
		ProjectPaths: []string{dir},
		UserPaths:    []string{},
	}

	skills, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("skills = %v", skills)
	}
	if skills[0].Name() != "found" {
		t.Fatalf("name = %q", skills[0].Name())
	}
}
