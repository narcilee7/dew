package storage

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/narcilee7/dew/pkg/event"
	"github.com/narcilee7/dew/pkg/fs"
	"github.com/narcilee7/dew/pkg/ai"
	"github.com/narcilee7/dew/pkg/session"
)

// SessionMeta stores session metadata.
type SessionMeta struct {
	ID        string            `json:"id"`
	ParentID  string            `json:"parent_id,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

// SessionState stores the compacted base state of a session.
type SessionState struct {
	ID       string       `json:"id"`
	Messages []ai.Message `json:"messages"`
}

// JSONLSessionStore persists sessions as event-sourced files.
type JSONLSessionStore struct {
	store *FileStore
}

// NewJSONLSessionStore creates a new JSONL-backed session store.
func NewJSONLSessionStore(fsys fs.FileSystem) *JSONLSessionStore {
	return &JSONLSessionStore{store: NewFileStore(fsys)}
}

func (s *JSONLSessionStore) sessionDir(id string) string {
	return id
}

// Create initializes a new session directory.
func (s *JSONLSessionStore) Create(ctx context.Context, opts session.CreateOptions) (session.Session, error) {
	dir := s.sessionDir(opts.ID)
	exists, err := s.store.FileSystem().Exists(ctx, dir)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("session %q already exists", opts.ID)
	}

	meta := SessionMeta{
		ID:        opts.ID,
		ParentID:  opts.ParentID,
		Metadata:  opts.Metadata,
		CreatedAt: time.Now(),
	}
	if err := s.writeMeta(ctx, opts.ID, meta); err != nil {
		return nil, err
	}
	if err := s.writeState(ctx, opts.ID, SessionState{ID: opts.ID}); err != nil {
		return nil, err
	}

	return session.NewMemorySession(opts.ID, opts.ParentID, opts.FS, opts.Sandbox), nil
}

// Load reads a session from disk.
func (s *JSONLSessionStore) Load(ctx context.Context, id string) (session.Session, error) {
	return s.resume(ctx, id, false)
}

// Resume loads a session and replays events newer than the checkpoint.
func (s *JSONLSessionStore) Resume(ctx context.Context, id string) (session.Session, error) {
	return s.resume(ctx, id, true)
}

func (s *JSONLSessionStore) resume(ctx context.Context, id string, replay bool) (session.Session, error) {
	meta, err := s.readMeta(ctx, id)
	if err != nil {
		return nil, err
	}

	// TODO: restore FS and Sandbox from meta/options.
	sess := session.NewMemorySession(meta.ID, meta.ParentID, nil, nil)

	state, err := s.readState(ctx, id)
	if err == nil {
		for _, m := range state.Messages {
			_ = sess.Append(ctx, m)
		}
	}

	if replay {
		// TODO: replay events from events/ directory.
	}

	return sess, nil
}

func (s *JSONLSessionStore) writeMeta(ctx context.Context, id string, meta SessionMeta) error {
	return s.store.WriteJSON(ctx, filepath.Join(s.sessionDir(id), "meta.json"), meta)
}

func (s *JSONLSessionStore) readMeta(ctx context.Context, id string) (SessionMeta, error) {
	var meta SessionMeta
	if err := s.store.ReadJSON(ctx, filepath.Join(s.sessionDir(id), "meta.json"), &meta); err != nil {
		return SessionMeta{}, err
	}
	return meta, nil
}

func (s *JSONLSessionStore) writeState(ctx context.Context, id string, state SessionState) error {
	return s.store.WriteJSON(ctx, filepath.Join(s.sessionDir(id), "state.json"), state)
}

func (s *JSONLSessionStore) readState(ctx context.Context, id string) (SessionState, error) {
	var state SessionState
	if err := s.store.ReadJSON(ctx, filepath.Join(s.sessionDir(id), "state.json"), &state); err != nil {
		return SessionState{}, err
	}
	return state, nil
}

// AppendEvent writes an event to the session event log.
func (s *JSONLSessionStore) AppendEvent(ctx context.Context, id string, ev event.Event) error {
	// TODO: serialize event and append to events/<seq>.jsonl
	_ = ev
	_ = ctx
	_ = id
	return nil
}

