package provider

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/sashabaranov/go-openai"
	"github/zqr233qr/story-trim/internal/core/port"
)

type openAIProvider struct {
	client *openai.Client
	model  string
}

func NewOpenAIProvider(baseURL, apiKey, model string) *openAIProvider {
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}
	client := openai.NewClientWithConfig(config)
	return &openAIProvider{
		client: client,
		model:  model,
	}
}

// ChatStream 实现模式1：交互式流式输出
func (p *openAIProvider) ChatStream(ctx context.Context, system, user string) (<-chan string, error) {
	req := openai.ChatCompletionRequest{
		Model: p.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: system},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
		Stream: true,
	}

	stream, err := p.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, err
	}

	ch := make(chan string)
	go func() {
		defer stream.Close()
		defer close(ch)

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				return
			}

			if len(response.Choices) > 0 {
				content := response.Choices[0].Delta.Content
				if content != "" {
					select {
					case <-ctx.Done():
						return
					case ch <- content:
					}
				}
			}
		}
	}()

	return ch, nil
}

// ChatJSON 实现模式2：后台任务结构化返回 (Legacy JSON Mode)
func (p *openAIProvider) ChatJSON(ctx context.Context, system, user string) (*port.BatchResult, error) {
	req := openai.ChatCompletionRequest{
		Model: p.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: system},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	}

	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil { return nil, err }

	if len(resp.Choices) == 0 { return nil, errors.New("empty response from llm") }

	var res port.BatchResult
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Chat 实现模式3：普通文本对话
func (p *openAIProvider) Chat(ctx context.Context, system, user string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: p.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: system},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
	}

	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil { return "", err }

	if len(resp.Choices) == 0 { return "", errors.New("empty response from llm") }

	return resp.Choices[0].Message.Content, nil
}