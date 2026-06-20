package ai

import (
	"context"
	"encoding/json"
)

// ResponseFunc is a function that produces a response from context.
type ResponseFunc func(ctx context.Context, model Model, context Context, opts ChatOptions) (*Response, error)

// MockProvider is a deterministic provider for testing.
type MockProvider struct {
	Response *Response
	Fn       ResponseFunc
}

// NewMockProvider creates a mock provider with a fixed response.
func NewMockProvider(resp *Response) *MockProvider {
	return &MockProvider{Response: resp}
}

// NewMockProviderFunc creates a mock provider with a response function.
func NewMockProviderFunc(fn ResponseFunc) *MockProvider {
	return &MockProvider{Fn: fn}
}

func (m *MockProvider) resolve(ctx context.Context, model Model, context Context, opts ChatOptions) (*Response, error) {
	if m.Fn != nil {
		return m.Fn(ctx, model, context, opts)
	}
	if m.Response == nil {
		return &Response{}, nil
	}
	return m.Response, nil
}

// Chat returns the mock response.
func (m *MockProvider) Chat(ctx context.Context, model Model, context Context, opts ChatOptions) (Response, error) {
	resp, err := m.resolve(ctx, model, context, opts)
	if err != nil {
		return Response{}, err
	}
	return *resp, nil
}

// Embed returns empty embedding vectors.
func (m *MockProvider) Embed(ctx context.Context, model Model, inputs []string, opts EmbedOptions) ([][]float32, error) {
	out := make([][]float32, len(inputs))
	return out, nil
}

// MockToolCall is a helper to build a tool call.
func MockToolCall(id, name string, args map[string]any) ToolCall {
	b, _ := json.Marshal(args)
	return ToolCall{ID: id, Name: name, Arguments: b}
}
