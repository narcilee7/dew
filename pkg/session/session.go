package session

import (
	"context"
	"sync"

	"github.com/narcilee7/dew/pkg/fs"
	"github.com/narcilee7/dew/pkg/llm"
	"github.com/narcilee7/dew/pkg/sandbox"
)

// CreateOptions configures a new session.
type CreateOptions struct {
	ID       string
	ParentID string
	Metadata map[string]string
	FS       fs.FileSystem
	Sandbox  sandbox.Sandbox
}

// memorySession is an in-memory session implementation.
type memorySession struct {
	id       string
	parentID string
	messages []llm.Message
	fs       fs.FileSystem
	sandbox  sandbox.Sandbox
	mu       sync.RWMutex
}

// newMemorySession creates a new in-memory session.
func newMemorySession(id, parentID string, fsys fs.FileSystem, box sandbox.Sandbox) *memorySession {
	return &memorySession{
		id:       id,
		parentID: parentID,
		messages: make([]llm.Message, 0),
		fs:       fsys,
		sandbox:  box,
	}
}

// NewMemorySession creates a new in-memory session and returns it as the Session interface.
func NewMemorySession(id, parentID string, fsys fs.FileSystem, box sandbox.Sandbox) Session {
	return newMemorySession(id, parentID, fsys, box)
}

// ID returns the session identifier.
func (s *memorySession) ID() string { return s.id }

// ParentID returns the parent session identifier.
func (s *memorySession) ParentID() string { return s.parentID }

// Messages returns a copy of the session messages.
func (s *memorySession) Messages() []llm.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]llm.Message, len(s.messages))
	copy(out, s.messages)
	return out
}

// Append adds a message to the session.
func (s *memorySession) Append(ctx context.Context, msg llm.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, msg)
	return nil
}

// Fork creates a child session.
func (s *memorySession) Fork(ctx context.Context, id string) (Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	child := newMemorySession(id, s.id, s.fs, s.sandbox)
	child.messages = make([]llm.Message, len(s.messages))
	copy(child.messages, s.messages)
	return child, nil
}

// Compact is a no-op for in-memory sessions.
func (s *memorySession) Compact(ctx context.Context) error {
	// TODO: implement summarization
	return nil
}

// FileSystem returns the session filesystem.
func (s *memorySession) FileSystem() fs.FileSystem { return s.fs }

// Sandbox returns the session sandbox.
func (s *memorySession) Sandbox() sandbox.Sandbox { return s.sandbox }

// Close releases session resources.
func (s *memorySession) Close() error {
	if s.fs != nil {
		_ = s.fs.Close()
	}
	if s.sandbox != nil {
		_ = s.sandbox.Close()
	}
	return nil
}
