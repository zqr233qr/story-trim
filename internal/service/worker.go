package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github/zqr233qr/story-trim/internal/domain"
)

type WorkerService struct {
	llm  LLMProcessor
	book *BookService
}

func NewWorkerService(llm LLMProcessor, book *BookService) *WorkerService {
	return &WorkerService{
		llm:  llm,
		book: book,
	}
}

// GenerateSummary 异步生成高质量摘要
func (w *WorkerService) GenerateSummary(content string, md5 string, version string) {
	// 使用一个独立的 Context，防止父请求取消导致中断
	ctx := context.Background()
	
	prompt := "请阅读以下小说章节，用不超过 200 字简要概括本章发生的核心剧情、主要人物的状态变化（如等级提升、受伤）以及获得的关键道具或线索。请直接输出摘要，不要包含任何客套话。"
	
	// 调用非流式接口 (假设 LLMProcessor 有 TrimContent，这里我们复用它，或者新增一个 GenerateSummary 接口)
	// 为了简单，我们暂时复用 TrimContent，虽然名字叫 Trim，但只要 Prompt 变了，它就是生成摘要。
	// 但更好的做法是 LLMService 暴露一个通用的 Chat 接口。
	// 这里我们暂时拼接 Prompt 并在 LLMService 内部处理，或者假设 TrimContent 只是一个单纯的 Chat 包装。
	// 修改策略：我们在 LLMService 增加一个 Chat 方法，或者直接用 TrimContent 传入特定 Prompt。
	
	// 临时方案：构造一个带指令的输入给 TrimContent (稍微有点 Hack，但能用)
	input := fmt.Sprintf("System: %s\n\nUser: %s", prompt, content)
	summary, err := w.llm.TrimContent(input)
	if err != nil {
		log.Error().Err(err).Str("md5", md5).Msg("Failed to generate summary")
		return
	}

	if err := w.book.SaveSummary(md5, summary, version); err != nil {
		log.Error().Err(err).Msg("Failed to save summary")
	} else {
		log.Info().Str("md5", md5).Msg("Summary generated and saved")
	}
}

// CheckAndGenerateEncyclopedia 检查并生成百科
func (w *WorkerService) CheckAndGenerateEncyclopedia(bookID uint, currentIdx int) {
	// 检查是否达到触发点 (例如每 50 章)
	interval := 50
	if currentIdx > 0 && currentIdx%interval == 0 {
		w.runEncyclopediaTask(bookID, currentIdx, interval)
	}
}

func (w *WorkerService) runEncyclopediaTask(bookID uint, endIdx int, rangeSize int) {
	ctx := context.Background()
	
	// 1. 获取书籍信息
	book, err := w.book.GetBook(bookID)
	if err != nil || book.Fingerprint == "" {
		return
	}

	// 2. 获取旧百科 (RangeEnd = endIdx - rangeSize)
	oldContext, _ := w.book.GetRelevantEncyclopedia(book.Fingerprint, endIdx) // 这里逻辑稍微调整，获取的是 < endIdx 的

	// 3. 获取最近区间的摘要
	startIdx := endIdx - rangeSize + 1
	summaries, err := w.book.GetPreviousSummaries(bookID, endIdx+1, rangeSize) // 获取该区间内的所有摘要
	if err != nil || len(summaries) == 0 {
		return
	}

	// 4. 调用 LLM 合并
	prompt := `你是一个文学设定分析员。请基于[旧百科]和[最新剧情摘要]，合并更新为一份最新的Markdown格式书籍设定集。
要求：
1. 更新人物状态（生死、等级、位置）。
2. 更新人际关系网。
3. 记录当前核心任务目标。
4. 移除已废弃的支线设定。
5. 总字数控制在 1000 字以内。`

	input := fmt.Sprintf("System: %s\n\n[旧百科]\n%s\n\n[最新剧情摘要]\n%s", 
		prompt, oldContext, strings.Join(summaries, "\n"))

	newEncyclopedia, err := w.llm.TrimContent(input)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate encyclopedia")
		return
	}

	// 5. 保存公共百科
	if err := w.book.SaveEncyclopedia(book.Fingerprint, endIdx, newEncyclopedia); err != nil {
		log.Error().Err(err).Msg("Failed to save encyclopedia")
	} else {
		log.Info().Str("book", book.Title).Int("range", endIdx).Msg("Encyclopedia updated")
	}
}
