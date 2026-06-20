package soul

import (
	"context"
	"time"

	"github.com/narcilee7/dew/pkg/event"
	"github.com/narcilee7/dew/pkg/memory"
)

// Soul is the persistent self-model of the agent.
type Soul struct {
	Identity     Identity     `json:"identity"`
	UserModel    UserModel    `json:"user_model"`
	Habits       Habits       `json:"habits"`
	GrowthGoals  []string     `json:"growth_goals"`
	LastUpdated  time.Time    `json:"last_updated"`
}

// Identity describes who the agent is.
type Identity struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Values      []string `json:"values"`
}

// UserModel describes what the agent has learned about the user.
type UserModel struct {
	Preferences []string `json:"preferences"`
	Patterns    []string `json:"patterns"`
	Notes       []string `json:"notes"`
}

// Habits describes the agent's working habits.
type Habits struct {
	PlanningStyle string   `json:"planning_style"`
	ExplainLevel  string   `json:"explain_level"`
	Rules         []string `json:"rules"`
}

// Update is a candidate change to the Soul.
type Update struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	SessionID   string    `json:"session_id"`
	Type        string    `json:"type"`
	Signal      string    `json:"signal"`
	Proposed    string    `json:"proposed"`
	Confidence  float64   `json:"confidence"`
	AutoApply   bool      `json:"auto_apply"`
}

// Observer watches the event stream and extracts candidate Soul updates.
type Observer interface {
	Observe(ctx context.Context, sessionID string, ev event.Event) (*Update, error)
	Flush(ctx context.Context, sessionID string) ([]Update, error)
}

// Reflector clusters updates and proposes changes to the Soul.
type Reflector interface {
	Reflect(ctx context.Context, updates []Update) ([]Update, error)
}

// PersonaMixer selects and formats the Soul fragment for a task.
type PersonaMixer interface {
	Mix(ctx context.Context, soul Soul, task string) (string, error)
}

// Engine coordinates Soul observation, reflection, and mixing.
type Engine interface {
	Observer
	Reflector
	PersonaMixer
	Load(ctx context.Context) (Soul, error)
	Save(ctx context.Context, soul Soul) error
	ApplyUpdates(ctx context.Context, soul Soul, updates []Update) (Soul, []Update, error)
}

// ToMemory converts a Soul update into a memory entry.
func (u Update) ToMemory() memory.Memory {
	return memory.Memory{
		Type:      memory.MemoryTypeSoul,
		Content:   u.Proposed,
		Source:    u.SessionID,
		CreatedAt: u.Timestamp,
	}
}
