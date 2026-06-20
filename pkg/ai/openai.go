package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// OpenAIProvider calls the OpenAI API via the official Go SDK.
type OpenAIProvider struct {
	client *openai.Client
}

// NewOpenAIProvider creates an OpenAI provider with the given API key.
func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	return &OpenAIProvider{client: openai.NewClient(apiKey)}
}

// NewOpenAIProviderWithBaseURL creates an OpenAI provider with a custom base URL.
// This is useful for OpenRouter, Azure OpenAI, or other OpenAI-compatible endpoints.
func NewOpenAIProviderWithBaseURL(apiKey, baseURL string) *OpenAIProvider {
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = baseURL
	return &OpenAIProvider{client: openai.NewClientWithConfig(cfg)}
}

// Chat sends a chat request to OpenAI.
func (p *OpenAIProvider) Chat(ctx context.Context, model Model, context Context, opts ChatOptions) (Response, error) {
	if model == nil {
		return Response{}, fmt.Errorf("model is nil")
	}

	req := openai.ChatCompletionRequest{
		Model:    model.ID(),
		Messages: toOpenAIMessages(context),
		Tools:    toOpenAITools(context.Tools),
	}
	if opts.MaxTokens > 0 {
		req.MaxTokens = opts.MaxTokens
	}
	if opts.Temperature != 0 {
		req.Temperature = float32(opts.Temperature)
	}

	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return Response{}, fmt.Errorf("openai chat: %w", err)
	}

	if len(resp.Choices) == 0 {
		return Response{}, fmt.Errorf("no choices in openai response")
	}
	choice := resp.Choices[0]

	return Response{
		Content:   choice.Message.Content,
		ToolCalls: fromOpenAIToolCalls(choice.Message.ToolCalls),
	}, nil
}

// Embed returns embedding vectors for the given inputs.
func (p *OpenAIProvider) Embed(ctx context.Context, model Model, inputs []string, opts EmbedOptions) ([][]float32, error) {
	if model == nil {
		return nil, fmt.Errorf("model is nil")
	}

	resp, err := p.client.CreateEmbeddings(ctx, openai.EmbeddingRequestStrings{
		Input: inputs,
		Model: openai.EmbeddingModel(model.ID()),
	})
	if err != nil {
		return nil, fmt.Errorf("openai embeddings: %w", err)
	}

	out := make([][]float32, len(resp.Data))
	for i, d := range resp.Data {
		out[i] = d.Embedding
	}
	return out, nil
}

func toOpenAIMessages(context Context) []openai.ChatCompletionMessage {
	msgs := make([]openai.ChatCompletionMessage, 0, len(context.Messages)+1)
	if context.SystemPrompt != "" {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: context.SystemPrompt,
		})
	}
	for _, m := range context.Messages {
		om := openai.ChatCompletionMessage{
			Role:       m.Role,
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
			Name:       m.Name,
		}
		for _, tc := range m.ToolCalls {
			args, _ := tc.Arguments.MarshalJSON()
			om.ToolCalls = append(om.ToolCalls, openai.ToolCall{
				ID:   tc.ID,
				Type: openai.ToolTypeFunction,
				Function: openai.FunctionCall{
					Name:      tc.Name,
					Arguments: string(args),
				},
			})
		}
		msgs = append(msgs, om)
	}
	return msgs
}

func toOpenAITools(tools []Tool) []openai.Tool {
	if len(tools) == 0 {
		return nil
	}
	out := make([]openai.Tool, len(tools))
	for i, t := range tools {
		out[i] = openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		}
	}
	return out
}

func fromOpenAIToolCalls(calls []openai.ToolCall) []ToolCall {
	if len(calls) == 0 {
		return nil
	}
	out := make([]ToolCall, len(calls))
	for i, c := range calls {
		out[i] = ToolCall{
			ID:        c.ID,
			Name:      c.Function.Name,
			Arguments: json.RawMessage(c.Function.Arguments),
		}
	}
	return out
}
