package config

// Config is the top-level dew configuration.
type Config struct {
	Agent    AgentConfig    `toml:"agent"`
	Server   ServerConfig   `toml:"server"`
	Pool     PoolConfig     `toml:"pool"`
	Registry RegistryConfig `toml:"registry"`
	MCP      MCPConfig      `toml:"mcp"`
	A2A      A2AConfig      `toml:"a2a"`
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
	MaxIdle int           `toml:"max_idle"`
	TTL     int           `toml:"ttl_ms"`
	Warm    []WarmSpec    `toml:"warm"`
}

// WarmSpec describes a pre-warmed agent spec.
type WarmSpec struct {
	Capability string `toml:"capability"`
	Count      int    `toml:"count"`
	Model      string `toml:"model,omitempty"`
}

// RegistryConfig configures the agent registry.
type RegistryConfig struct {
	Type    string            `toml:"type"` // "memory", "etcd", "consul"
	Address string            `toml:"address,omitempty"`
	Params  map[string]string `toml:"params,omitempty"`
}

// MCPConfig configures the MCP server.
type MCPConfig struct {
	Enabled bool   `toml:"enabled"`
	Transport string `toml:"transport"` // "stdio", "sse"
}

// A2AConfig configures the A2A server.
type A2AConfig struct {
	Enabled bool   `toml:"enabled"`
	Address string `toml:"address"`
}

// Default returns a default configuration.
func Default() Config {
	return Config{
		Agent: AgentConfig{
			Model:    "openai:gpt-4o",
			MaxTurns: 50,
			Timeout:  300_000, // 5 minutes
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
	}
}
