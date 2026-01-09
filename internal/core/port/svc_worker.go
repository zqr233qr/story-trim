package port

import (
	"context"
)

type WorkerService interface {
	// SubmitFullTrimTask 提交全书精简异步任务
	SubmitFullTrimTask(ctx context.Context, userID uint, bookID uint, promptID uint) (string, error)
	// GenerateSummary 内部调用：生成章节摘要
	GenerateSummary(ctx context.Context, bookFP string, index int, md5 string, content string)
	// UpdateEncyclopedia 内部调用：更新书籍百科
	UpdateEncyclopedia(ctx context.Context, bookFP string, endIdx int)
}
