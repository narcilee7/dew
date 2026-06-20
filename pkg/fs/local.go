package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type local struct {
	root string
}

func NewLocal(root string) FileSystem {
	return &local{root: root}
}

func (l *local) Root() string {
	return l.root
}

func (l *local) resolve(path string) (string, error) {
	if path == "" {
		return l.root, nil
	}
	clean := filepath.Clean(path)
	if strings.Contains(clean, "..") {
		return "", fmt.Errorf("path %q contains ..", path)
	}
	abs := filepath.Join(l.root, clean)
	if !strings.HasPrefix(abs, l.root) {
		return "", fmt.Errorf("path %q is outside root", path)
	}
	return abs, nil
}

func (l *local) Read(ctx context.Context, path string) ([]byte, error) {
	abs, err := l.resolve(path)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(abs)
}

// Write writes a file, creating parent directories as needed.
func (l *local) Write(ctx context.Context, path string, data []byte) error {
	abs, err := l.resolve(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	return os.WriteFile(abs, data, 0o644)
}

// Delete removes a file.
func (l *local) Delete(ctx context.Context, path string) error {
	abs, err := l.resolve(path)
	if err != nil {
		return err
	}
	return os.Remove(abs)
}

// Exists reports whether a path exists.
func (l *local) Exists(ctx context.Context, path string) (bool, error) {
	abs, err := l.resolve(path)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(abs)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// List returns directory entries.
func (l *local) List(ctx context.Context, path string) ([]FileInfo, error) {
	abs, err := l.resolve(path)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, err
	}
	infos := make([]FileInfo, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		infos = append(infos, FileInfo{
			Name:  e.Name(),
			Path:  filepath.Join(path, e.Name()),
			IsDir: e.IsDir(),
			Size:  info.Size(),
		})
	}
	return infos, nil
}

func (l *local) Close() error {
	if l.root != "" {
		l.root = ""
		return nil
	}
	return fmt.Errorf("already closed")
}
