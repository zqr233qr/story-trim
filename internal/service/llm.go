package service

import (
	"context"

	"github.com/sashabaranov/go-openai"
	"github.com/zqr233qr/story-trim/internal/config"
)

type LlmService struct {
	llm *config.LLM
}

func NewLlmService(llm *config.LLM) *LlmService {
	return &LlmService{
		llm: llm,
	}
}

// 百万
var million = 1000000.0

func (s *LlmService) getLlmConfig() *config.LLMConfig {
	llmConfig := s.llm.LLMConfig[s.llm.Use]
	llmConfig.InputPrice = llmConfig.InputPrice * 100
	llmConfig.OutputPrice = llmConfig.OutputPrice * 100
	return &llmConfig
}

func (s *LlmService) Llm(ctx context.Context, systemPrompt string, userPrompt string) (*LlmResponse, error) {
	llmConfig := s.getLlmConfig()
	conf := openai.DefaultConfig(llmConfig.APIKey)
	conf.BaseURL = llmConfig.BaseURL
	client := openai.NewClientWithConfig(conf)

	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: llmConfig.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	llmResponse := &LlmResponse{
		LlmName:           s.llm.Use,
		Resp:              &resp,
		TotalTokens:       resp.Usage.TotalTokens,
		PromptTokens:      resp.Usage.PromptTokens,
		CompletionTokens:  resp.Usage.CompletionTokens,
		InputMTokenPrice:  llmConfig.InputPrice,
		OutputMTokenPrice: llmConfig.OutputPrice,
		InputCost:         float64(resp.Usage.PromptTokens) * llmConfig.InputPrice / million,
		OutputCost:        float64(resp.Usage.CompletionTokens) * llmConfig.OutputPrice / million,
	}

	llmResponse.TotalCost = llmResponse.InputCost + llmResponse.OutputCost

	return llmResponse, nil
}

func (s *LlmService) LlmWithStream(ctx context.Context, systemPrompt string, userPrompt string) (*LlmResponse, error) {
	llmConfig := s.getLlmConfig()
	conf := openai.DefaultConfig(llmConfig.APIKey)
	conf.BaseURL = llmConfig.BaseURL
	client := openai.NewClientWithConfig(conf)

	stream, err := client.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model: llmConfig.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
			Stream: true,
			// 关键点：开启流式输出的 Usage 统计
			StreamOptions: &openai.StreamOptions{
				IncludeUsage: true,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	llmResponse := &LlmResponse{
		LlmName:           s.llm.Use,
		Stream:            stream,
		InputMTokenPrice:  llmConfig.InputPrice,
		OutputMTokenPrice: llmConfig.OutputPrice,
	}

	return llmResponse, nil
}

type LlmResponse struct {
	LlmName           string
	Resp              *openai.ChatCompletionResponse
	Stream            *openai.ChatCompletionStream
	TotalTokens       int
	PromptTokens      int
	CompletionTokens  int
	InputMTokenPrice  float64
	OutputMTokenPrice float64
	InputCost         float64
	OutputCost        float64
	TotalCost         float64
}

type LlmServiceInterface interface {
	Llm(ctx context.Context, systemPrompt string, userPrompt string) (*LlmResponse, error)
	LlmWithStream(ctx context.Context, systemPrompt string, userPrompt string) (*LlmResponse, error)
}
