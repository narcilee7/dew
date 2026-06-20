package log

import (
	"context"
	"fmt"
	"time"

	"github.com/narcilee7/dew/pkg/fs"
	"github.com/narcilee7/dew/pkg/storage"
)

// FileAuditLogger writes audit records as JSON Lines to dated files.
type FileAuditLogger struct {
	store *storage.FileStore
}

// NewFileAuditLogger creates a new file-based audit logger rooted at the given filesystem.
func NewFileAuditLogger(fsys fs.FileSystem) *FileAuditLogger {
	return &FileAuditLogger{store: storage.NewFileStore(fsys)}
}

func (l *FileAuditLogger) path() string {
	return fmt.Sprintf("%s.jsonl", time.Now().UTC().Format("2006-01-02"))
}

// ToolCall records a tool invocation.
func (l *FileAuditLogger) ToolCall(ctx context.Context, record ToolCallRecord) error {
	return l.store.AppendJSONL(ctx, l.path(), record)
}

// SafetyDecision records a safety decision.
func (l *FileAuditLogger) SafetyDecision(ctx context.Context, record SafetyRecord) error {
	return l.store.AppendJSONL(ctx, l.path(), record)
}

// Delegation records a subagent delegation.
func (l *FileAuditLogger) Delegation(ctx context.Context, record DelegationRecord) error {
	return l.store.AppendJSONL(ctx, l.path(), record)
}

// FileAccess records a filesystem access.
func (l *FileAuditLogger) FileAccess(ctx context.Context, record FileAccessRecord) error {
	return l.store.AppendJSONL(ctx, l.path(), record)
}
