package config

import "context"

// Scope identifies the configuration lifetime.
type Scope int

const (
	// ScopeUser is the user's global configuration.
	ScopeUser Scope = iota
	// ScopeProject is the project-level configuration.
	ScopeProject
	// ScopeSession is the configuration overlay for a single session.
	ScopeSession
)

// Config is the top-level dew configuration.
type Config struct {
	Agent    AgentConfig    `toml:"agent"`
	Server   ServerConfig   `toml:"server"`
	Pool     PoolConfig     `toml:"pool"`
	Registry RegistryConfig `toml:"registry"`
	MCP      MCPConfig      `toml:"mcp"`
	A2A      A2AConfig      `toml:"a2a"`
	Soul     SoulConfig     `toml:"soul"`
}

// AgentConfig configures the core agent.
type AgentConfig struct {
	Model        string   `toml:"model"`
	SystemPrompt string   `toml:"system_prompt"`
	MaxTurns     int      `toml:"max_turns"`
	Timeout      int      `toml:"timeout_ms"`
	Tools        []string `toml:"tools"`
}

// ServerConfig configures the gRPC server.
type ServerConfig struct {
	Address string `toml:"address"`
	TLS     bool   `toml:"tls"`
}

// PoolConfig configures the agent pool.
type PoolConfig struct {
	MaxIdle int        `toml:"max_idle"`
	TTL     int        `toml:"ttl_ms"`
	Warm    []WarmSpec `toml:"warm"`
}

// WarmSpec describes a pre-warmed agent spec.
type WarmSpec struct {
	Capability string `toml:"capability"`
	Count      int    `toml:"count"`
	Model      string `toml:"model,omitempty"`
}

// RegistryConfig configures the agent registry.
type RegistryConfig struct {
	Type    string            `toml:"type"`
	Address string            `toml:"address,omitempty"`
	Params  map[string]string `toml:"params,omitempty"`
}

// MCPConfig configures the MCP server.
type MCPConfig struct {
	Enabled   bool   `toml:"enabled"`
	Transport string `toml:"transport"`
}

// A2AConfig configures the A2A server.
type A2AConfig struct {
	Enabled bool   `toml:"enabled"`
	Address string `toml:"address"`
}

// SoulConfig configures the SOUL layer.
type SoulConfig struct {
	Enabled      bool     `toml:"enabled"`
	AutoUpdate   bool     `toml:"auto_update"`
	Path         string   `toml:"path,omitempty"`
	Persona      string   `toml:"persona,omitempty"`
	ConfirmRisky bool     `toml:"confirm_risky"`
}

// Store loads, saves, and watches configuration.
type Store interface {
	Load(ctx context.Context, scope Scope) (Config, error)
	Save(ctx context.Context, scope Scope, cfg Config) error
	Watch(ctx context.Context, scope Scope) (<-chan Config, error)
}

// Default returns a default configuration.
func Default() Config {
	return Config{
		Agent: AgentConfig{
			Model:    "openai:gpt-4o",
			MaxTurns: 50,
			Timeout:  300_000,
			Tools:    []string{"read", "bash", "task"},
		},
		Server: ServerConfig{
			Address: "unix:/var/run/dew/dew.sock",
		},
		Pool: PoolConfig{
			MaxIdle: 4,
			TTL:     300_000,
		},
		Registry: RegistryConfig{
			Type: "memory",
		},
		MCP: MCPConfig{
			Enabled:   true,
			Transport: "stdio",
		},
		A2A: A2AConfig{
			Enabled: false,
		},
		Soul: SoulConfig{
			Enabled:      true,
			AutoUpdate:   true,
			ConfirmRisky: true,
		},
	}
}
