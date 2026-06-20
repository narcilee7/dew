package memory

import (
	"context"
	"time"
)

// MemoryType classifies a memory entry.
type MemoryType string

const (
	MemoryTypeRule       MemoryType = "rule"
	MemoryTypeFact       MemoryType = "fact"
	MemoryTypePreference MemoryType = "preference"
	MemoryTypeLesson     MemoryType = "lesson"
	MemoryTypeSoul       MemoryType = "soul"
)

// Memory is a single persisted piece of knowledge.
type Memory struct {
	ID        string     `json:"id"`
	Type      MemoryType `json:"type"`
	Content   string     `json:"content"`
	Source    string     `json:"source,omitempty"`
	Project   string     `json:"project,omitempty"`
	Embedding []float32  `json:"embedding,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// RecallQuery queries the memory store for relevant entries.
type RecallQuery struct {
	Query   string
	Project string
	Types   []MemoryType
	Limit   int
}

// ListFilter filters memories for listing.
type ListFilter struct {
	Project string
	Types   []MemoryType
	Limit   int
}

// Store persists and retrieves cross-session knowledge.
type Store interface {
	Recall(ctx context.Context, query RecallQuery) ([]Memory, error)
	Remember(ctx context.Context, mem Memory) error
	Forget(ctx context.Context, id string) error
	List(ctx context.Context, filter ListFilter) ([]Memory, error)
}
