package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/narcilee7/dew/pkg/agent"
	"github.com/narcilee7/dew/pkg/ai"
	"github.com/narcilee7/dew/pkg/core"
	"github.com/narcilee7/dew/pkg/fs"
	"github.com/narcilee7/dew/pkg/memory"
	"github.com/narcilee7/dew/pkg/plan"
	"github.com/narcilee7/dew/pkg/sandbox"
	"github.com/narcilee7/dew/pkg/session"
	"github.com/narcilee7/dew/pkg/skill"
	"github.com/narcilee7/dew/pkg/soul"
	"github.com/narcilee7/dew/pkg/tools"
	"github.com/narcilee7/dew/pkg/trajectory"
)

// Runtime holds the runtime dependencies for CLI commands.
// It is distinct from core.Harness: Runtime is the CLI-level wiring,
// while core.Harness is the agent runtime itself.
type Runtime struct {
	Logger   *slog.Logger
	Config   Config
	Provider ai.Provider
	Registry tools.ToolRegistry
	Session  session.Store
	Factory  agent.AgentFactory
	Pool     agent.Pool
	Harness  core.Harness
}

// Config is a simplified runtime config.
type Config struct {
	Model        string
	SystemPrompt string
	MaxTurns     int
	TimeoutMs    int
	Tools        []string
	DataDir      string
}

// DefaultConfig returns the default CLI config.
func DefaultConfig() Config {
	home, _ := os.UserHomeDir()
	return Config{
		Model:        "openai:gpt-4o",
		SystemPrompt: "You are dew, a helpful coding agent.",
		MaxTurns:     50,
		TimeoutMs:    300_000,
		Tools:        []string{"read", "bash"},
		DataDir:      filepath.Join(home, ".local", "share", "dew"),
	}
}

// NewRuntime builds the CLI runtime.
func NewRuntime(cfg Config) (*Runtime, error) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	var provider ai.Provider
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey != "" {
		provider = ai.NewOpenAIProvider(apiKey)
		logger.Info("using OpenAI provider")
	} else {
		provider = ai.NewMockProviderFunc(mockProvider)
		logger.Info("using mock provider; set OPENAI_API_KEY for real models")
	}

	registry := core.NewToolRegistry()
	_ = registry.Register(&tools.ReadTool{})
	_ = registry.Register(&tools.BashTool{})
	_ = registry.Register(&agent.TaskTool{Pool: agent.NewPool(nil, agent.PoolOptions{})})

	sessStore := session.NewMemoryStore()

	factory := agent.NewLocalFactory(provider, sessStore, registry, logger)
	pool := agent.NewPool(factory, agent.PoolOptions{MaxIdle: 4})

	harness := core.NewHarness("", core.Boundaries{
		Provider: provider,
		Tools:    registry,
		Session:  sessStore,
		Logger:   logger,
	})
	harness.SetLoop(&core.DefaultLoop{
		Context: &core.SimpleContextManager{SystemPrompt: cfg.SystemPrompt},
		Logger:  logger,
	})

	// Load project-local and user-global skills.
	loader := skill.NewLoader()
	skills, err := loader.Load(context.Background())
	if err != nil {
		logger.Warn("failed to load skills", "error", err)
	} else {
		for _, s := range skills {
			if err := harness.Use(s.Plugin()); err != nil {
				logger.Warn("failed to use skill", "skill", s.Name(), "error", err)
			} else {
				logger.Info("loaded skill", "skill", s.Name())
			}
		}
	}

	// Register built-in memory / soul / plan plugins backed by persistent storage.
	if err := os.MkdirAll(cfg.DataDir, 0o755); err == nil {
		dataFS := fs.NewLocal(cfg.DataDir)
		_ = harness.Use(memory.NewPlugin(dataFS))
		_ = harness.Use(soul.NewPlugin(dataFS, "soul/soul.md"))
		_ = harness.Use(plan.NewPlugin(dataFS))
		_ = harness.Use(trajectory.NewPlugin(dataFS))
		logger.Info("loaded built-in plugins", "datadir", cfg.DataDir)
	} else {
		logger.Warn("failed to create data dir, built-in plugins disabled", "error", err)
	}

	return &Runtime{
		Logger:   logger,
		Config:   cfg,
		Provider: provider,
		Registry: registry,
		Session:  sessStore,
		Factory:  factory,
		Pool:     pool,
		Harness:  harness,
	}, nil
}

// NewSession creates a new session with isolated filesystem and sandbox.
func (r *Runtime) NewSession(ctx context.Context) (session.Session, error) {
	root, err := os.MkdirTemp("", "dew-*")
	if err != nil {
		return nil, fmt.Errorf("create root: %w", err)
	}
	fsys := fs.NewLocal(root)
	box := sandbox.NewLocal("main", fsys)

	return r.Session.Create(ctx, session.CreateOptions{
		ID:      fmt.Sprintf("session-%d", time.Now().UnixNano()),
		FS:      fsys,
		Sandbox: box,
	})
}

// mockProvider is a simple provider that echoes back a completion.
func mockProvider(ctx context.Context, model ai.Model, context ai.Context, opts ai.ChatOptions) (*ai.Response, error) {
	for i := len(context.Messages) - 1; i >= 0; i-- {
		m := context.Messages[i]
		if m.Role == core.RoleAssistant && len(m.ToolCalls) > 0 {
			return &ai.Response{Content: "Done."}, nil
		}
		if m.Role == core.RoleUser {
			break
		}
	}
	return &ai.Response{
		Content: "I'll run a command for you.",
		ToolCalls: []ai.ToolCall{
			ai.MockToolCall("call-1", "bash", map[string]any{"command": "echo hello from dew"}),
		},
	}, nil
}
