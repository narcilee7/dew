package soul

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/narcilee7/dew/pkg/fs"
	"github.com/narcilee7/dew/pkg/storage"
)

// defaultSoul returns a fresh default Soul.
func defaultSoul() Soul {
	return Soul{
		Identity: Identity{
			Name:        "dew",
			Description: "A coding assistant that values clarity over cleverness.",
			Values: []string{
				"prefer small, reviewable changes",
				"explain intent before risky edits",
				"ask before irreversible operations",
			},
		},
		UserModel: UserModel{},
		Habits: Habits{
			PlanningStyle: "start with trade-offs",
			ExplainLevel:  "balanced",
			Rules: []string{
				"write tests when fixing or adding behavior",
				"avoid heavy dependencies for trivial utilities",
			},
		},
		GrowthGoals: []string{
			"learn to write shorter explanations",
			"remember to ask before editing config files",
		},
		LastUpdated: time.Now(),
	}
}

// FileStore loads and saves the Soul from a Markdown file.
type FileStore struct {
	store *storage.FileStore
	path  string
}

// NewFileStore creates a new file-based Soul store rooted at the given filesystem.
func NewFileStore(fsys fs.FileSystem, path string) *FileStore {
	return &FileStore{
		store: storage.NewFileStore(fsys),
		path:  path,
	}
}

// Load reads the Soul from disk. Missing files return the default Soul.
func (s *FileStore) Load(ctx context.Context) (Soul, error) {
	data, err := s.store.ReadString(ctx, s.path)
	if err != nil {
		return defaultSoul(), nil
	}
	return parseSoulMarkdown(string(data))
}

// Save writes the Soul to disk as Markdown.
func (s *FileStore) Save(ctx context.Context, soul Soul) error {
	soul.LastUpdated = time.Now()
	return s.store.WriteString(ctx, s.path, renderSoulMarkdown(soul))
}

func parseSoulMarkdown(text string) (Soul, error) {
	// TODO: parse markdown sections into Soul struct.
	_ = text
	return defaultSoul(), nil
}

func renderSoulMarkdown(soul Soul) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# dew SOUL\n\n")
	fmt.Fprintf(&b, "## Identity\n%s\n\n", soul.Identity.Description)
	for _, v := range soul.Identity.Values {
		fmt.Fprintf(&b, "- %s\n", v)
	}
	fmt.Fprintf(&b, "\n## User Model\n")
	for _, p := range soul.UserModel.Preferences {
		fmt.Fprintf(&b, "- %s\n", p)
	}
	fmt.Fprintf(&b, "\n## Habits\n")
	fmt.Fprintf(&b, "- Planning style: %s\n", soul.Habits.PlanningStyle)
	fmt.Fprintf(&b, "- Explain level: %s\n", soul.Habits.ExplainLevel)
	for _, r := range soul.Habits.Rules {
		fmt.Fprintf(&b, "- %s\n", r)
	}
	fmt.Fprintf(&b, "\n## Growth Goals\n")
	for _, g := range soul.GrowthGoals {
		fmt.Fprintf(&b, "- %s\n", g)
	}
	fmt.Fprintf(&b, "\n<!-- last_updated: %s -->\n", soul.LastUpdated.Format(time.RFC3339))
	return b.String()
}
