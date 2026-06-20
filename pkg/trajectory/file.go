package trajectory

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/narcilee7/dew/pkg/fs"
	"github.com/narcilee7/dew/pkg/storage"
)

// FileStore is a file-backed trajectory store.
type FileStore struct {
	store *storage.FileStore
}

// NewFileStore creates a trajectory store rooted at the given filesystem.
func NewFileStore(fsys fs.FileSystem) *FileStore {
	return &FileStore{store: storage.NewFileStore(fsys)}
}

func (s *FileStore) trajDir(id string) string {
	return id
}

func (s *FileStore) metaPath(id string) string {
	return filepath.Join(s.trajDir(id), "trajectory.json")
}

func (s *FileStore) eventsPath(id string) string {
	return filepath.Join(s.trajDir(id), "events.jsonl")
}

// Create writes a new trajectory metadata file.
func (s *FileStore) Create(ctx context.Context, traj Trajectory) error {
	if traj.CreatedAt.IsZero() {
		traj.CreatedAt = time.Now()
	}
	if traj.UpdatedAt.IsZero() {
		traj.UpdatedAt = traj.CreatedAt
	}
	return s.store.WriteJSON(ctx, s.metaPath(traj.ID), traj)
}

// Save updates trajectory metadata.
func (s *FileStore) Save(ctx context.Context, traj Trajectory) error {
	traj.UpdatedAt = time.Now()
	return s.store.WriteJSON(ctx, s.metaPath(traj.ID), traj)
}

// Load reads trajectory metadata.
func (s *FileStore) Load(ctx context.Context, id string) (Trajectory, error) {
	var traj Trajectory
	if err := s.store.ReadJSON(ctx, s.metaPath(id), &traj); err != nil {
		return Trajectory{}, err
	}
	return traj, nil
}

// AppendEvent appends an event record to the trajectory's JSONL log.
func (s *FileStore) AppendEvent(ctx context.Context, trajectoryID string, rec EventRecord) error {
	return s.store.AppendJSONL(ctx, s.eventsPath(trajectoryID), rec)
}

// ListEvents reads all event records for a trajectory.
func (s *FileStore) ListEvents(ctx context.Context, trajectoryID string) ([]EventRecord, error) {
	data, err := s.store.ReadString(ctx, s.eventsPath(trajectoryID))
	if err != nil {
		return nil, err
	}

	var out []EventRecord
	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var rec EventRecord
		// Unmarshal payload lazily as raw message so concrete event types can be
		// decoded later.
		if err := json.Unmarshal(line, &rec); err != nil {
			return nil, fmt.Errorf("decode event: %w", err)
		}
		out = append(out, rec)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

