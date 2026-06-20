package llm

import (
	"context"
	"encoding/json"
)

// ResponseFunc is a function that produces a response from context.
type ResponseFunc func(ctx context.Context, model Model, context Context, opts Options) (*Response, error)

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

func (m *MockProvider) resolve(ctx context.Context, model Model, context Context, opts Options) (*Response, error) {
	if m.Fn != nil {
		return m.Fn(ctx, model, context, opts)
	}
	if m.Response == nil {
		return &Response{}, nil
	}
	return m.Response, nil
}

// Stream returns a single response as a stream.
func (m *MockProvider) Stream(ctx context.Context, model Model, context Context, opts Options) (EventStream, error) {
	resp, err := m.resolve(ctx, model, context, opts)
	if err != nil {
		return nil, err
	}

	ch := make(chan Event, 10)
	go func() {
		defer close(ch)
		ch <- TextStartEvent{}
		ch <- TextDeltaEvent{Delta: resp.Content}
		ch <- TextEndEvent{}
		for _, tc := range resp.ToolCalls {
			ch <- ToolCallStartEvent{ID: tc.ID, Name: tc.Name}
			ch <- ToolCallDeltaEvent{ID: tc.ID, Arguments: string(tc.Arguments)}
			ch <- ToolCallEndEvent{ID: tc.ID}
		}
		ch <- DoneEvent{}
	}()
	return &mockStream{ch: ch}, nil
}

// Complete returns the response.
func (m *MockProvider) Complete(ctx context.Context, model Model, context Context, opts Options) (Response, error) {
	resp, err := m.resolve(ctx, model, context, opts)
	if err != nil {
		return Response{}, err
	}
	return *resp, nil
}

type mockStream struct {
	ch chan Event
}

func (s *mockStream) Recv() (Event, error) {
	ev, ok := <-s.ch
	if !ok {
		return nil, context.Canceled
	}
	return ev, nil
}

func (s *mockStream) Close() error { return nil }

// MockToolCall is a helper to build a tool call.
func MockToolCall(id, name string, args map[string]any) ToolCall {
	b, _ := json.Marshal(args)
	return ToolCall{ID: id, Name: name, Arguments: b}
}
