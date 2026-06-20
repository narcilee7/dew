package ai

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIProviderChat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{
			"choices": [{
				"message": {
					"role": "assistant",
					"content": "hello from openai",
					"tool_calls": [{
						"id": "call-1",
						"type": "function",
						"function": {"name": "bash", "arguments": "{\"command\":\"echo hi\"}"}
					}]
				}
			}]
		}`)
	}))
	defer server.Close()

	provider := NewOpenAIProviderWithBaseURL("test-key", server.URL)
	model := NewModel("gpt-4o", ModelCapabilities{Chat: true, Tools: true})

	resp, err := provider.Chat(context.Background(), model, Context{
		SystemPrompt: "You are helpful.",
		Messages: []Message{
			{Role: "user", Content: "say hi"},
		},
		Tools: []Tool{
			{Name: "bash", Description: "run shell commands", Parameters: map[string]any{}},
		},
	}, ChatOptions{})

	if err != nil {
		t.Fatalf("chat: %v", err)
	}
	if resp.Content != "hello from openai" {
		t.Fatalf("content = %q", resp.Content)
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("tool calls = %v", resp.ToolCalls)
	}
	if resp.ToolCalls[0].Name != "bash" {
		t.Fatalf("tool name = %q", resp.ToolCalls[0].Name)
	}
}

func TestOpenAIProviderEmbed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embeddings" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{
			"data": [
				{"embedding": [0.1, 0.2, 0.3]},
				{"embedding": [0.4, 0.5, 0.6]}
			]
		}`)
	}))
	defer server.Close()

	provider := NewOpenAIProviderWithBaseURL("test-key", server.URL)
	model := NewModel("text-embedding-3-small", ModelCapabilities{Embedding: true})

	out, err := provider.Embed(context.Background(), model, []string{"hello", "world"}, EmbedOptions{})
	if err != nil {
		t.Fatalf("embed: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("embeddings = %v", out)
	}
	if len(out[0]) != 3 {
		t.Fatalf("first embedding = %v", out[0])
	}
}

func TestOpenAIProviderMissingKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"message":"Unauthorized"}}`, http.StatusUnauthorized)
	}))
	defer server.Close()

	provider := NewOpenAIProviderWithBaseURL("", server.URL)
	model := NewModel("gpt-4o", ModelCapabilities{Chat: true})
	_, err := provider.Chat(context.Background(), model, Context{}, ChatOptions{})
	if err == nil {
		t.Fatal("expected error for missing api key")
	}
}
