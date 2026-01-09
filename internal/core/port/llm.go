package port

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type BatchResult struct {
	TrimmedContent string `json:"trimmed_content"`
	Summary        string `json:"summary"`
}

type LLMPort interface {
	// ChatStream 交互式流式精简 (模式1)
	ChatStream(ctx context.Context, system, user string) (<-chan string, *openai.Usage, error)

	// ChatJSON 后台任务结构化返回 (模式2)
	ChatJSON(ctx context.Context, system, user string) (*BatchResult, *openai.Usage, error)

	// Chat 基础文本对话 (模式3 - 用于 XML 模式或通用对话)
	Chat(ctx context.Context, system, user string) (string, *openai.Usage, error)
}

type StoragePort interface {
	Save(ctx context.Context, filename string, data []byte) (string, error)
	Delete(ctx context.Context, path string) error
}
