package port

import (
	"context"
)

type BatchResult struct {
	TrimmedContent string `json:"trimmed_content"`
	Summary        string `json:"summary"`
}

type LLMPort interface {
	// ChatStream 交互式流式精简 (模式1)
	ChatStream(ctx context.Context, system, user string) (<-chan string, error)
	
	// ChatJSON 后台任务结构化返回 (模式2)
	ChatJSON(ctx context.Context, system, user string) (*BatchResult, error)

	// Chat 基础文本对话 (模式3 - 用于 XML 模式或通用对话)
	Chat(ctx context.Context, system, user string) (string, error)
}

type StoragePort interface {
	Save(ctx context.Context, filename string, data []byte) (string, error)
	Delete(ctx context.Context, path string) error
}
