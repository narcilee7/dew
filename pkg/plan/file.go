package plan

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/narcilee7/dew/pkg/fs"
	"github.com/narcilee7/dew/pkg/storage"
)

// FileStore is a file-based plan store.
type FileStore struct {
	store *storage.FileStore
}

// NewFileStore creates a new file-based plan store rooted at the given filesystem.
func NewFileStore(fsys fs.FileSystem) *FileStore {
	return &FileStore{store: storage.NewFileStore(fsys)}
}

func (s *FileStore) planDir(id string) string {
	return id
}

func (s *FileStore) checkpointDir(planID string) string {
	return filepath.Join(s.planDir(planID), "checkpoints")
}

// Create writes a new plan to disk.
func (s *FileStore) Create(ctx context.Context, plan Plan) error {
	return s.store.WriteJSON(ctx, filepath.Join(s.planDir(plan.ID), "plan.json"), plan)
}

// Load reads a plan from disk.
func (s *FileStore) Load(ctx context.Context, id string) (Plan, error) {
	var plan Plan
	if err := s.store.ReadJSON(ctx, filepath.Join(s.planDir(id), "plan.json"), &plan); err != nil {
		return Plan{}, err
	}
	return plan, nil
}

// SaveCheckpoint writes a checkpoint.
func (s *FileStore) SaveCheckpoint(ctx context.Context, planID string, cp Checkpoint) error {
	return s.store.WriteJSON(ctx, filepath.Join(s.checkpointDir(planID), fmt.Sprintf("%s.json", cp.ID)), cp)
}

// LoadCheckpoint reads a checkpoint.
func (s *FileStore) LoadCheckpoint(ctx context.Context, planID string, checkpointID string) (Checkpoint, error) {
	var cp Checkpoint
	if err := s.store.ReadJSON(ctx, filepath.Join(s.checkpointDir(planID), fmt.Sprintf("%s.json", checkpointID)), &cp); err != nil {
		return Checkpoint{}, err
	}
	return cp, nil
}

// ListCheckpoints returns all checkpoints for a plan.
func (s *FileStore) ListCheckpoints(ctx context.Context, planID string) ([]CheckpointInfo, error) {
	names, err := s.store.ListJSON(ctx, s.checkpointDir(planID))
	if err != nil {
		return nil, err
	}

	var out []CheckpointInfo
	for _, name := range names {
		var cp Checkpoint
		if err := s.store.ReadJSON(ctx, filepath.Join(s.checkpointDir(planID), fmt.Sprintf("%s.json", name)), &cp); err != nil {
			continue
		}
		out = append(out, CheckpointInfo{
			ID:        cp.ID,
			Label:     cp.Label,
			CreatedAt: cp.CreatedAt,
		})
	}
	return out, nil
}
