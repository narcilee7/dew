package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/narcilee7/dew/pkg/fs"
)

// Command describes a command to execute inside a sandbox.
type Command struct {
	Args    []string
	Env     map[string]string
	WorkDir string
	Stdin   []byte
	Timeout time.Duration
}

// ExecResult describes the outcome of a sandboxed command.
type ExecResult struct {
	ExitCode int
	Stdout   []byte
	Stderr   []byte
}

// Sandbox is an isolated execution environment.
type Sandbox interface {
	io.Closer

	// ID returns the sandbox identifier.
	ID() string

	// Exec runs a command inside the sandbox.
	Exec(ctx context.Context, cmd Command) (ExecResult, error)

	// FileSystem returns the sandbox's filesystem view.
	FileSystem() fs.FileSystem
}

// Provider creates sandbox instances.
type Provider interface {
	Create(ctx context.Context, opts CreateOptions) (Sandbox, error)
}

// CreateOptions configures a new sandbox.
type CreateOptions struct {
	ID        string
	Root      string
	ReadOnly  bool
	Network   NetworkPolicy
	Resources ResourceLimits
}

// NetworkPolicy controls sandbox network access.
type NetworkPolicy struct {
	DefaultDeny bool
	Allowlist   []string
}

// ResourceLimits controls sandbox resource usage.
type ResourceLimits struct {
	CPUs   float64
	Memory int64 // bytes
	Disk   int64 // bytes
}

// Local is a local-process sandbox that runs commands in a restricted directory.
type Local struct {
	id   string
	fsys fs.FileSystem
}

// NewLocal creates a new local sandbox.
func NewLocal(id string, fsys fs.FileSystem) *Local {
	return &Local{id: id, fsys: fsys}
}

// ID returns the sandbox identifier.
func (l *Local) ID() string { return l.id }

// FileSystem returns the sandbox filesystem.
func (l *Local) FileSystem() fs.FileSystem { return l.fsys }

// Close is a no-op for local sandbox.
func (l *Local) Close() error { return nil }

// Exec runs a command locally with restricted working directory.
func (l *Local) Exec(ctx context.Context, cmd Command) (ExecResult, error) {
	if len(cmd.Args) == 0 {
		return ExecResult{}, fmt.Errorf("command args are empty")
	}

	workDir := l.fsys.Root()
	if cmd.WorkDir != "" {
		abs, err := filepath.Abs(filepath.Join(workDir, cmd.WorkDir))
		if err != nil {
			return ExecResult{}, err
		}
		if !isSubpath(abs, workDir) {
			return ExecResult{}, fmt.Errorf("workdir %q escapes sandbox root", cmd.WorkDir)
		}
		workDir = abs
	}

	if cmd.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cmd.Timeout)
		defer cancel()
	}

	c := exec.CommandContext(ctx, cmd.Args[0], cmd.Args[1:]...)
	c.Dir = workDir
	c.Env = mergeEnv(os.Environ(), cmd.Env)
	if len(cmd.Stdin) > 0 {
		c.Stdin = bytes.NewReader(cmd.Stdin)
	}

	stdout, err := c.Output()
	res := ExecResult{
		ExitCode: 0,
		Stdout:   stdout,
	}
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			res.ExitCode = ee.ExitCode()
			res.Stderr = ee.Stderr
		} else {
			res.ExitCode = 1
			res.Stderr = []byte(err.Error())
		}
	}
	return res, nil
}

func isSubpath(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel != ".." && !filepath.IsAbs(rel) && !strings.HasPrefix(rel, "..")
}

func mergeEnv(base []string, extra map[string]string) []string {
	if len(extra) == 0 {
		return base
	}
	// Build a map from base, then override with extra.
	outMap := make(map[string]string, len(base)+len(extra))
	for _, e := range base {
		if i := strings.IndexByte(e, '='); i >= 0 {
			outMap[e[:i]] = e[i+1:]
		}
	}
	for k, v := range extra {
		outMap[k] = v
	}
	out := make([]string, 0, len(outMap))
	for k, v := range outMap {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}
	return out
}
