package session

import (
	"context"
	"fmt"
	"sync"
)

// memoryStore is an in-memory implementation of Store.
type memoryStore struct {
	sessions map[string]*memorySession
	mu       sync.RWMutex
}

// NewMemoryStore creates a new in-memory session store.
func NewMemoryStore() Store {
	return &memoryStore{sessions: make(map[string]*memorySession)}
}

// Create creates a new session.
func (st *memoryStore) Create(ctx context.Context, opts CreateOptions) (Session, error) {
	st.mu.Lock()
	defer st.mu.Unlock()
	if _, exists := st.sessions[opts.ID]; exists {
		return nil, fmt.Errorf("session %q already exists", opts.ID)
	}
	sess := newMemorySession(opts.ID, opts.ParentID, opts.FS, opts.Sandbox)
	st.sessions[opts.ID] = sess
	return sess, nil
}

// Load loads an existing session.
func (st *memoryStore) Load(ctx context.Context, id string) (Session, error) {
	st.mu.RLock()
	defer st.mu.RUnlock()
	sess, ok := st.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session %q not found", id)
	}
	return sess, nil
}

// Resume reactivates an existing session. For the in-memory store it is equivalent to Load.
func (st *memoryStore) Resume(ctx context.Context, id string) (Session, error) {
	// TODO: distinguish resume from load once sessions support persistence / hibernation.
	return st.Load(ctx, id)
}
