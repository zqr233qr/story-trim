package service

import (
	"context"
	"github/zqr233qr/story-trim/internal/domain"
)

// Splitter 定义分章服务的接口
type Splitter interface {
	// SplitFile 读取文件并进行分章
	SplitFile(filePath string) ([]domain.SplitChapter, error)
	// SplitContent 直接处理文本内容
	SplitContent(content string) []domain.SplitChapter
}

// LLMProcessor 定义 AI 处理服务的接口
type LLMProcessor interface {
	// TrimContent 接收原始文本，返回精简后的文本
	TrimContent(content string) (string, error)
	// TrimContentStream 返回一个 channel 用于流式传输增量内容
	TrimContentStream(ctx context.Context, content string) (<-chan string, error)
}
