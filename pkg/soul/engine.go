package soul

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/narcilee7/dew/pkg/event"
)

// SimpleEngine is a minimal SOUL engine implementation.
type SimpleEngine struct {
	store   Store
	buffer  []Update
	maxAuto float64
}

// Store persists and loads Soul.
type Store interface {
	Load(ctx context.Context) (Soul, error)
	Save(ctx context.Context, soul Soul) error
}

// NewSimpleEngine creates a new SimpleEngine.
func NewSimpleEngine(store Store) *SimpleEngine {
	return &SimpleEngine{
		store:   store,
		maxAuto: 0.7,
	}
}

// Load reads the Soul from disk.
func (e *SimpleEngine) Load(ctx context.Context) (Soul, error) {
	return e.store.Load(ctx)
}

// Save writes the Soul to disk.
func (e *SimpleEngine) Save(ctx context.Context, soul Soul) error {
	return e.store.Save(ctx, soul)
}

// Observe watches a single event and may produce an update.
func (e *SimpleEngine) Observe(ctx context.Context, sessionID string, ev event.Event) (*Update, error) {
	// TODO: inspect event types and generate updates.
	_ = ev
	return nil, nil
}

// Flush returns all buffered updates and clears the buffer.
func (e *SimpleEngine) Flush(ctx context.Context, sessionID string) ([]Update, error) {
	out := e.buffer
	e.buffer = nil
	return out, nil
}

// Reflect clusters updates and filters low-confidence ones.
func (e *SimpleEngine) Reflect(ctx context.Context, updates []Update) ([]Update, error) {
	// TODO: cluster similar updates and boost confidence.
	return updates, nil
}

// Mix formats a Soul fragment for injection into a prompt.
func (e *SimpleEngine) Mix(ctx context.Context, soul Soul, task string) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "You are %s. %s\n\n", soul.Identity.Name, soul.Identity.Description)
	if len(soul.Identity.Values) > 0 {
		fmt.Fprintf(&b, "Core values:\n")
		for _, v := range soul.Identity.Values {
			fmt.Fprintf(&b, "- %s\n", v)
		}
	}
	if len(soul.UserModel.Preferences) > 0 {
		fmt.Fprintf(&b, "\nUser preferences:\n")
		for _, p := range soul.UserModel.Preferences {
			fmt.Fprintf(&b, "- %s\n", p)
		}
	}
	if len(soul.Habits.Rules) > 0 {
		fmt.Fprintf(&b, "\nWorking habits:\n")
		for _, r := range soul.Habits.Rules {
			fmt.Fprintf(&b, "- %s\n", r)
		}
	}
	fmt.Fprintf(&b, "\nCurrent task: %s\n", task)
	return b.String(), nil
}

// ApplyUpdates merges approved updates into the Soul.
func (e *SimpleEngine) ApplyUpdates(ctx context.Context, soul Soul, updates []Update) (Soul, []Update, error) {
	var applied []Update
	for _, u := range updates {
		if u.Confidence >= e.maxAuto && u.AutoApply {
			soul = applyUpdate(soul, u)
			u.Timestamp = time.Now()
			applied = append(applied, u)
		}
	}
	return soul, applied, nil
}

func applyUpdate(soul Soul, u Update) Soul {
	switch u.Type {
	case "preference":
		soul.UserModel.Preferences = append(soul.UserModel.Preferences, u.Proposed)
	case "habit":
		soul.Habits.Rules = append(soul.Habits.Rules, u.Proposed)
	case "goal":
		soul.GrowthGoals = append(soul.GrowthGoals, u.Proposed)
	case "identity":
		soul.Identity.Description = u.Proposed
	}
	soul.LastUpdated = time.Now()
	return soul
}
