package config

import (
	"context"
	"fmt"

	"github.com/narcilee7/dew/pkg/fs"
	"github.com/narcilee7/dew/pkg/storage"
)

// FileStore loads and saves configuration from TOML files.
type FileStore struct {
	user    *storage.FileStore
	project *storage.FileStore
}

// NewFileStore creates a new file-based config store.
// userFS is rooted at the user config directory; projectFS at the project directory.
func NewFileStore(userFS, projectFS fs.FileSystem) *FileStore {
	return &FileStore{
		user:    storage.NewFileStore(userFS),
		project: storage.NewFileStore(projectFS),
	}
}

func (s *FileStore) storeForScope(scope Scope) *storage.FileStore {
	switch scope {
	case ScopeUser:
		return s.user
	case ScopeProject, ScopeSession:
		return s.project
	default:
		return nil
	}
}

func (s *FileStore) pathForScope(scope Scope) string {
	switch scope {
	case ScopeUser, ScopeProject:
		return "config.toml"
	case ScopeSession:
		return ".session/config.toml"
	default:
		return ""
	}
}

// Load reads configuration from disk. Missing files return the default config.
func (s *FileStore) Load(ctx context.Context, scope Scope) (Config, error) {
	cfg := Default()
	store := s.storeForScope(scope)
	path := s.pathForScope(scope)
	if store == nil || path == "" {
		return cfg, fmt.Errorf("unknown scope %d", scope)
	}

	data, err := store.ReadString(ctx, path)
	if err != nil {
		return cfg, nil
	}

	// TODO: decode TOML into cfg.
	_ = data

	return cfg, nil
}

// Save writes configuration to disk.
func (s *FileStore) Save(ctx context.Context, scope Scope, cfg Config) error {
	store := s.storeForScope(scope)
	path := s.pathForScope(scope)
	if store == nil || path == "" {
		return fmt.Errorf("unknown scope %d", scope)
	}

	// TODO: encode cfg to TOML.
	_ = cfg
	return store.WriteString(ctx, path, "")
}

// Watch returns a channel that emits configuration updates.
// Not implemented: returns a channel that never fires.
func (s *FileStore) Watch(ctx context.Context, scope Scope) (<-chan Config, error) {
	out := make(chan Config)
	go func() {
		<-ctx.Done()
		close(out)
	}()
	return out, nil
}
