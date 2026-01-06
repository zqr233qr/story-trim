package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github/zqr233qr/story-trim/internal/core/domain"
	"github/zqr233qr/story-trim/internal/core/port"
)

type workerService struct {
	bookRepo   port.BookRepository
	cacheRepo  port.CacheRepository
	actionRepo port.ActionRepository
	taskRepo   port.TaskRepository
	promptRepo port.PromptRepository
	llm        port.LLMPort
	cfg        *TrimConfig
}

func NewWorkerService(br port.BookRepository, cr port.CacheRepository, ar port.ActionRepository, tr port.TaskRepository, pr port.PromptRepository, llm port.LLMPort, cfg *TrimConfig) *workerService {
	return &workerService{
		bookRepo:   br,
		cacheRepo:  cr,
		actionRepo: ar,
		taskRepo:   tr,
		promptRepo: pr,
		llm:        llm,
		cfg:        cfg,
	}
}

func (s *workerService) GenerateSummary(ctx context.Context, bookFP string, index int, md5 string, content string) {
	jsonPrompt := "请概括本章剧情（200字内）。返回JSON: {\"summary\": \"你的摘要...\"}"
	res, err := s.llm.ChatJSON(ctx, jsonPrompt, content)
	if err != nil { return }

	summary := res.Summary
	if summary == "" { summary = res.TrimmedContent }

	_ = s.cacheRepo.SaveSummary(ctx, &domain.RawSummary{
		BookFingerprint: bookFP, ChapterIndex: index, Content: summary, Version: "v1.0", CreatedAt: time.Now(),
	})
}

func (s *workerService) StartBatchTrim(ctx context.Context, userID uint, bookID uint, promptID uint) (string, error) {
	taskID := uuid.New().String()
	task := &domain.Task{
		ID: taskID, UserID: userID, BookID: bookID, Type: "batch_trim", Status: "pending", Progress: 0, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := s.taskRepo.CreateTask(ctx, task); err != nil { return "", err }
	go s.runBatchTrimTask(context.Background(), task, promptID)
	return taskID, nil
}

func (s *workerService) GetTaskStatus(ctx context.Context, taskID string) (*domain.Task, error) {
	return s.taskRepo.GetTaskByID(ctx, taskID)
}

func (s *workerService) runBatchTrimTask(ctx context.Context, task *domain.Task, promptID uint) {
	task.Status = "running"
	_ = s.taskRepo.UpdateTask(ctx, task)

	book, _ := s.bookRepo.GetBookByID(ctx, task.BookID)
	chapters, _ := s.bookRepo.GetChaptersByBookID(ctx, task.BookID)
	prompt, _ := s.promptRepo.GetPromptByID(ctx, promptID)

	for i, chap := range chapters {
		cache, err := s.cacheRepo.GetTrimResult(ctx, chap.ContentMD5, promptID, prompt.Version)
		if err == nil && cache != nil && cache.Level == 2 {
			_ = s.actionRepo.RecordUserTrim(ctx, &domain.UserProcessedChapter{
				UserID: task.UserID, BookID: book.ID, ChapterID: chap.ID, PromptID: promptID, CreatedAt: time.Now(),
			})
		} else {
			s.processSingleBatchChapter(ctx, task.UserID, book, chap, prompt, 2)
		}
		task.Progress = int(float64(i+1) / float64(len(chapters)) * 100)
		task.UpdatedAt = time.Now()
		_ = s.taskRepo.UpdateTask(ctx, task)
		if (i+1)%s.cfg.EncyclopediaInterval == 0 {
			s.UpdateEncyclopedia(ctx, book.Fingerprint, i+1)
		}
	}
	task.Status = "completed"
	_ = s.taskRepo.UpdateTask(ctx, task)
}

func (s *workerService) processSingleBatchChapter(ctx context.Context, userID uint, book *domain.Book, chap domain.Chapter, prompt *domain.Prompt, targetLevel int) {
	summaries, _ := s.cacheRepo.GetSummaries(ctx, book.Fingerprint, chap.Index, s.cfg.SummaryLimit)
	encyclopedia, _ := s.cacheRepo.GetEncyclopedia(ctx, book.Fingerprint, chap.Index)
	raw, _ := s.bookRepo.GetRawContent(ctx, chap.ContentMD5)

	rawLen := len([]rune(raw.Content))
	var minRate, maxRate float64
	switch prompt.ID {
	case 1: minRate, maxRate = 0.70, 0.85
	case 3: minRate, maxRate = 0.15, 0.35
	default: minRate, maxRate = 0.45, 0.65
	}
	minWords := int(float64(rawLen) * minRate)
	maxWords := int(float64(rawLen) * maxRate)

	template := strings.ReplaceAll(prompt.Content, "{MIN_WORDS}", fmt.Sprintf("%d", minWords))
	template = strings.ReplaceAll(template, "{MAX_WORDS}", fmt.Sprintf("%d", maxWords))

	var sb strings.Builder
	sb.WriteString("### [核心编辑协议]\n")
	sb.WriteString("- 对话红线：严禁改写或删减原文对话，必须原样保留对话字句。\n")
	sb.WriteString("- 格式要求：请严格按照 XML 标签输出，不要输出任何 Markdown 代码块。\n")
	sb.WriteString("  格式如下：\n")
	sb.WriteString("  <content>\n  (这里是精简后的小说正文)\n  </content>\n\n")
	sb.WriteString("  <summary>\n  (这里是本章剧情摘要)\n  </summary>\n")
	sb.WriteString("- 指标要求：")
	sb.WriteString(template)
	sb.WriteString("\n\n")
	
	if encyclopedia != nil { sb.WriteString("### [全局百科]\n" + encyclopedia.Content + "\n\n") }
	if len(summaries) > 0 {
		sb.WriteString("### [前情提要]\n")
		for _, sm := range summaries { sb.WriteString(sm.Content + "\n") }
	}

	log.Debug().Str("prompt", sb.String()).Int("idx", chap.Index).Msg("Batch Task: Assembled XML Prompt")
	
	// 使用纯文本 Chat 接口
	resText, err := s.llm.Chat(ctx, sb.String(), raw.Content)
	if err != nil {
		log.Error().Err(err).Int("idx", chap.Index).Msg("Batch Task: LLM Chat failed")
		return
	}

	// 手动解析 XML 标签
	trimmedContent := extractTagContent(resText, "content")
	summaryContent := extractTagContent(resText, "summary")

	// 容错：如果解析失败，但内容很长，假设全部都是正文
	if trimmedContent == "" {
		if len(resText) > len(raw.Content)/2 {
			trimmedContent = resText
		} else {
			log.Warn().Int("idx", chap.Index).Str("raw_response", resText).Msg("Batch Task: Failed to parse XML output")
			// 仍然保存，防止任务卡死，但标记一下？或者直接返回
			return 
		}
	}
	if summaryContent == "" { summaryContent = "摘要生成失败" }

	trimmedLen := len([]rune(trimmedContent))
	rate := 0.0
	if rawLen > 0 {
		rate = float64(trimmedLen) / float64(rawLen)
		rate = float64(int(rate*10000+0.5)) / 10000
	}

	_ = s.cacheRepo.SaveTrimResult(ctx, &domain.TrimResult{
		ContentMD5:     chap.ContentMD5,
		PromptID:       prompt.ID,
		PromptVersion:  prompt.Version,
		Level:          targetLevel,
		TrimmedContent: trimmedContent,
		TrimWords:      trimmedLen,
		TrimRate:       rate,
		CreatedAt:      time.Now(),
	})
	_ = s.cacheRepo.SaveSummary(ctx, &domain.RawSummary{
		BookFingerprint: book.Fingerprint, ChapterIndex: chap.Index, Content: summaryContent, Version: "v1.0", CreatedAt: time.Now(),
	})
	_ = s.actionRepo.RecordUserTrim(ctx, &domain.UserProcessedChapter{
		UserID: userID, BookID: book.ID, ChapterID: chap.ID, PromptID: prompt.ID, CreatedAt: time.Now(),
	})
}

func (s *workerService) UpdateEncyclopedia(ctx context.Context, bookFP string, endIdx int) {
	old, _ := s.cacheRepo.GetEncyclopedia(ctx, bookFP, endIdx)
	summaries, _ := s.cacheRepo.GetSummaries(ctx, bookFP, endIdx+1, s.cfg.EncyclopediaInterval)
	if len(summaries) == 0 { return }
	var summaryTexts []string
	for _, sm := range summaries { summaryTexts = append(summaryTexts, sm.Content) }
	prompt := "你是一个文学设定分析员。请基于[旧百科]和[最新剧情摘要]，合并更新为最新的Markdown设定集。"
	input := fmt.Sprintf("[旧百科]\n%s\n\n[新剧情摘要]\n%s", func() string { if old != nil { return old.Content }; return "暂无" }(), strings.Join(summaryTexts, "\n"))
	
	// 这里继续使用 ChatJSON，因为百科更新适合结构化思维
	res, err := s.llm.ChatJSON(ctx, prompt, input)
	if err != nil { return }
	
	// 注意：ChatJSON 返回 BatchResult，但我们只需要 Summary 字段（复用）或者 TrimmedContent
	content := res.Summary
	if content == "" { content = res.TrimmedContent }
	
	_ = s.cacheRepo.SaveEncyclopedia(ctx, &domain.SharedEncyclopedia{
		BookFingerprint: bookFP, RangeEnd: endIdx, Content: content, Version: "v1.0", CreatedAt: time.Now(),
	})
}

// 辅助函数：提取 XML 标签内容
func extractTagContent(text, tagName string) string {
	startTag := "<" + tagName + ">"
	endTag := "</" + tagName + ">"
	startIdx := strings.Index(text, startTag)
	endIdx := strings.LastIndex(text, endTag)

	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		return ""
	}
	return strings.TrimSpace(text[startIdx+len(startTag) : endIdx])
}
