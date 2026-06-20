package fs

import (
	"context"
	"io"
)

// FileSystem is an isolated view of a filesystem.
// All paths are relative to the boundary root and must not escape it.
type FileSystem interface {
	io.Closer

	// Read returns the contents of a file.
	Read(ctx context.Context, path string) ([]byte, error)

	// Write creates or overwrites a file.
	Write(ctx context.Context, path string, data []byte) error

	// Delete removes a file.
	Delete(ctx context.Context, path string) error

	// Exists reports whether a path exists.
	Exists(ctx context.Context, path string) (bool, error)

	// List returns entries in a directory.
	List(ctx context.Context, path string) ([]FileInfo, error)

	// Root returns the absolute root path of this filesystem view.
	Root() string
}

type FileInfo struct {
	Name  string
	Path  string
	IsDir bool
	Size  int64
}
