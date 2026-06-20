package storage

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/narcilee7/dew/pkg/fs"
)

// FileStore provides generic JSON/JSONL file operations on top of fs.FileSystem.
type FileStore struct {
	fsys fs.FileSystem
}

// NewFileStore creates a file store backed by the given filesystem.
func NewFileStore(fsys fs.FileSystem) *FileStore {
	return &FileStore{fsys: fsys}
}

// FileSystem returns the underlying filesystem.
func (s *FileStore) FileSystem() fs.FileSystem {
	return s.fsys
}

// ReadJSON reads and unmarshals a JSON file.
func (s *FileStore) ReadJSON(ctx context.Context, path string, v any) error {
	data, err := s.fsys.Read(ctx, path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// WriteJSON marshals and writes a JSON file.
func (s *FileStore) WriteJSON(ctx context.Context, path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return s.fsys.Write(ctx, path, data)
}

// AppendJSONL appends a single JSON line to a file.
func (s *FileStore) AppendJSONL(ctx context.Context, path string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.AppendBytes(ctx, path, append(data, '\n'))
}

// WriteBytes writes raw bytes.
func (s *FileStore) WriteBytes(ctx context.Context, path string, data []byte) error {
	return s.fsys.Write(ctx, path, data)
}

// AppendBytes appends raw bytes, creating the file if needed.
func (s *FileStore) AppendBytes(ctx context.Context, path string, data []byte) error {
	existing, err := s.fsys.Read(ctx, path)
	if err != nil {
		return s.fsys.Write(ctx, path, data)
	}
	return s.fsys.Write(ctx, path, append(existing, data...))
}

// ReadString reads a file as string.
func (s *FileStore) ReadString(ctx context.Context, path string) (string, error) {
	data, err := s.fsys.Read(ctx, path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteString writes a string.
func (s *FileStore) WriteString(ctx context.Context, path string, data string) error {
	return s.fsys.Write(ctx, path, []byte(data))
}

// Delete removes a file.
func (s *FileStore) Delete(ctx context.Context, path string) error {
	return s.fsys.Delete(ctx, path)
}

// ListJSON returns all .json file names in a directory (without extension).
func (s *FileStore) ListJSON(ctx context.Context, dir string) ([]string, error) {
	entries, err := s.fsys.List(ctx, dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir {
			continue
		}
		if filepath.Ext(e.Name) == ".json" {
			out = append(out, strings.TrimSuffix(e.Name, ".json"))
		}
	}
	return out, nil
}
