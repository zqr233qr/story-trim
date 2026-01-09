package port

import (
	"context"
)

type TrimService interface {
	// TrimStreamByMD5 基于内容哈希的流式精简逻辑 (适用于本地优先模式)
	TrimStreamByMD5(ctx context.Context, userID uint, chapterMD5 string, content string, promptID uint, chapterIndex int, bookFingerprint string) (<-chan string, error)
	// TrimStreamByChapterID 基于云端章节ID的流式精简逻辑 (适用于小程序/云端模式)
	TrimStreamByChapterID(ctx context.Context, userID uint, bookID uint, chapterID uint, promptID uint) (<-chan string, error)
}
