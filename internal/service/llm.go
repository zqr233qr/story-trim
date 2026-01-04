package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github/zqr233qr/story-trim/pkg/config"
	"github.com/rs/zerolog/log"
)

type LLMService struct {
	config config.LLMConfig
	client *http.Client
}

func NewLLMService(cfg config.LLMConfig) *LLMService {
	return &LLMService{
		config: cfg,
		client: &http.Client{
			// 对于流式请求，不应该设置过短的总超时，或者依靠 Request Context 来控制超时
			Timeout: 0, 
		},
	}
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type streamResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

const prompt = "你是一个专业的文学编辑。请阅读以下小说章节，保留所有对话、关键动作和剧情转折。请删除冗余的环境描写、重复的心理活动和无关的填充文字。保持原有的叙事风格。目标是将字数减少 30%-50%。请直接输出精简后的正文，不要包含任何解释语或Markdown标记。"

// TrimContent 一次性返回结果 (旧接口保留)
func (s *LLMService) TrimContent(content string) (string, error) {
	reqBody := chatRequest{
		Model: s.config.Model,
		Messages: []message{
			{Role: "system", Content: prompt},
			{Role: "user", Content: content},
		},
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// 临时创建一个带超时的 client 用于非流式
	client := &http.Client{Timeout: time.Duration(s.config.Timeout) * time.Second}
	req, err := http.NewRequest("POST", s.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	log.Debug().Str("model", s.config.Model).Msg("Sending request to LLM (Blocking)")
	
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("api returned error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from ai")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// TrimContentStream 流式处理实现
func (s *LLMService) TrimContentStream(ctx context.Context, content string) (<-chan string, error) {
	reqBody := chatRequest{
		Model: s.config.Model,
		Messages: []message{
			{Role: "system", Content: prompt},
			{Role: "user", Content: content},
		},
		Stream: true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)
	req.Header.Set("Accept", "text/event-stream")

	log.Debug().Msg("Sending request to LLM (Stream)")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(body))
	}

	ch := make(chan string)

	go func() {
		defer resp.Body.Close()
		defer close(ch)

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					log.Error().Err(err).Msg("Error reading stream")
				}
				return
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// SSE 格式：data: {json}
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}

			var streamResp streamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				log.Warn().Err(err).Str("line", line).Msg("Failed to unmarshal stream data")
				continue
			}

			if len(streamResp.Choices) > 0 {
				delta := streamResp.Choices[0].Delta.Content
				if delta != "" {
					// 检查 Context 是否被取消
					select {
					case <-ctx.Done():
						return
					case ch <- delta:
					}
				}
			}
		}
	}()

	return ch, nil
}