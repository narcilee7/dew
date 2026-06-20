package memory

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/narcilee7/dew/pkg/fs"
	"github.com/narcilee7/dew/pkg/storage"
)

// FileStore is a file-based memory store.
type FileStore struct {
	store *storage.FileStore
}

// NewFileStore creates a new file-based memory store rooted at the given filesystem.
func NewFileStore(fsys fs.FileSystem) *FileStore {
	return &FileStore{store: storage.NewFileStore(fsys)}
}

func (s *FileStore) path(id string) string {
	return fmt.Sprintf("%s.json", id)
}

// Remember writes a memory to disk.
func (s *FileStore) Remember(ctx context.Context, mem Memory) error {
	if mem.ID == "" {
		mem.ID = fmt.Sprintf("mem-%d", time.Now().UnixNano())
	}
	if mem.CreatedAt.IsZero() {
		mem.CreatedAt = time.Now()
	}
	return s.store.WriteJSON(ctx, s.path(mem.ID), mem)
}

// Forget removes a memory from disk.
func (s *FileStore) Forget(ctx context.Context, id string) error {
	return s.store.Delete(ctx, s.path(id))
}

// List returns memories matching the filter.
func (s *FileStore) List(ctx context.Context, filter ListFilter) ([]Memory, error) {
	entries, err := s.store.FileSystem().List(ctx, ".")
	if err != nil {
		return nil, err
	}

	typeSet := make(map[MemoryType]bool)
	for _, t := range filter.Types {
		typeSet[t] = true
	}

	var out []Memory
	for _, e := range entries {
		if e.IsDir || !strings.HasSuffix(e.Name, ".json") {
			continue
		}
		var mem Memory
		if err := s.store.ReadJSON(ctx, e.Name, &mem); err != nil {
			continue
		}
		if filter.Project != "" && mem.Project != filter.Project {
			continue
		}
		if len(typeSet) > 0 && !typeSet[mem.Type] {
			continue
		}
		out = append(out, mem)
		if filter.Limit > 0 && len(out) >= filter.Limit {
			break
		}
	}
	return out, nil
}

// Recall returns relevant memories for a query.
// TODO: implement semantic search / embedding retrieval.
func (s *FileStore) Recall(ctx context.Context, query RecallQuery) ([]Memory, error) {
	return s.List(ctx, ListFilter{
		Project: query.Project,
		Types:   query.Types,
		Limit:   query.Limit,
	})
}
