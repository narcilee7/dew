package session

import (
	"context"
	"io"

	"github.com/narcilee7/dew/pkg/fs"
	"github.com/narcilee7/dew/pkg/ai"
	"github.com/narcilee7/dew/pkg/sandbox"
)

// Session is an isolated conversation context.
type Session interface {
	io.Closer

	// ID returns the unique session identifier.
	ID() string

	// ParentID returns the parent session ID, if this session was forked.
	ParentID() string

	// Messages returns the current conversation messages.
	Messages() []ai.Message

	// Append adds a message to the session history.
	Append(ctx context.Context, msg ai.Message) error

	// Fork creates a child session with a copy of the current state.
	Fork(ctx context.Context, id string) (Session, error)

	// Compact reduces history size while preserving essential context.
	Compact(ctx context.Context) error

	// FileSystem returns the session's filesystem view.
	FileSystem() fs.FileSystem

	// Sandbox returns the session's sandbox.
	Sandbox() sandbox.Sandbox
}

// Store creates, loads, and resumes sessions.
type Store interface {
	Create(ctx context.Context, opts CreateOptions) (Session, error)
	Load(ctx context.Context, id string) (Session, error)
	Resume(ctx context.Context, id string) (Session, error)
}
