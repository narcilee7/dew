package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Model identifies a specific model and its capabilities.
// Providers work with Model instances rather than raw strings.
type Model interface {
	// ID returns the model identifier passed to the provider API.
	ID() string

	// Capabilities returns what this model can do.
	Capabilities() ModelCapabilities
}

// ModelCapabilities describes what a model can do.
type ModelCapabilities struct {
	Chat      bool
	Vision    bool
	Embedding bool
	Reasoning bool
	Tools     bool
	JSONMode  bool
}

// Has returns true if all listed capabilities are present.
func (c ModelCapabilities) Has(req ModelCapabilities) bool {
	if req.Chat && !c.Chat {
		return false
	}
	if req.Vision && !c.Vision {
		return false
	}
	if req.Embedding && !c.Embedding {
		return false
	}
	if req.Reasoning && !c.Reasoning {
		return false
	}
	if req.Tools && !c.Tools {
		return false
	}
	if req.JSONMode && !c.JSONMode {
		return false
	}
	return true
}

// staticModel is a simple Model implementation backed by a string id and capabilities.
type staticModel struct {
	id           string
	capabilities ModelCapabilities
}

// NewModel creates a Model from an id and capabilities.
func NewModel(id string, caps ModelCapabilities) Model {
	return &staticModel{id: id, capabilities: caps}
}

func (m *staticModel) ID() string                { return m.id }
func (m *staticModel) Capabilities() ModelCapabilities { return m.capabilities }

// ParseModel parses a model reference of the form "provider:model-id".
// If no provider prefix is present, it is treated as a raw id and the chat
// capability is enabled by default.
func ParseModel(ref string) (Model, error) {
	if ref == "" {
		return nil, fmt.Errorf("model reference is empty")
	}

	provider, id, ok := strings.Cut(ref, ":")
	if !ok {
		return NewModel(ref, ModelCapabilities{Chat: true}), nil
	}

	// Normalize known providers.
	switch provider {
	case "openai":
		return NewModel(id, ModelCapabilities{Chat: true, Tools: true, JSONMode: true}), nil
	case "anthropic":
		return NewModel(id, ModelCapabilities{Chat: true, Tools: true}), nil
	default:
		return NewModel(id, ModelCapabilities{Chat: true}), nil
	}
}

// Message is a single message in a conversation transcript.
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

// ToolCall represents a single tool invocation requested by the model.
type ToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// Tool describes a callable tool in the LLM context.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

// Context is the input to a chat call.
type Context struct {
	SystemPrompt string
	Messages     []Message
	Tools        []Tool
}

// Response is a normalized chat response.
type Response struct {
	Content   string
	ToolCalls []ToolCall
}

// ChatOptions configures a chat request.
type ChatOptions struct {
	APIKey      string
	BaseURL     string
	MaxTokens   int
	Temperature float64
}

// EmbedOptions configures an embedding request.
type EmbedOptions struct {
	APIKey  string
	BaseURL string
}

// Provider is the unified interface for all model backends.
// A provider may implement chat, embedding, vision, or any combination.
type Provider interface {
	// Chat sends a chat context to the model and returns a response.
	Chat(ctx context.Context, model Model, context Context, opts ChatOptions) (Response, error)

	// Embed returns embedding vectors for the given inputs.
	Embed(ctx context.Context, model Model, inputs []string, opts EmbedOptions) ([][]float32, error)
}

// ChatProvider is a provider that only supports chat.
type ChatProvider interface {
	Chat(ctx context.Context, model Model, context Context, opts ChatOptions) (Response, error)
}

// EmbedProvider is a provider that only supports embeddings.
type EmbedProvider interface {
	Embed(ctx context.Context, model Model, inputs []string, opts EmbedOptions) ([][]float32, error)
}
