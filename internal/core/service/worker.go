package service

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github/zqr233qr/story-trim/internal/core/domain"
	"github/zqr233qr/story-trim/internal/core/port"
	"github/zqr233qr/story-trim/pkg/config"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type workerService struct {
	bookRepo   port.BookRepository
	cacheRepo  port.CacheRepository
	actionRepo port.ActionRepository
	taskRepo   port.TaskRepository
	promptRepo port.PromptRepository
	llm        port.LLMPort
	cfg        *TrimConfig
	tmplBatch  *template.Template
	tmplSum    *template.Template
	tmplTrim   *template.Template
}

func NewWorkerService(br port.BookRepository, cr port.CacheRepository, ar port.ActionRepository, tr port.TaskRepository, pr port.PromptRepository, llm port.LLMPort, cfg *TrimConfig) *workerService {
	tmplBatch, _ := template.ParseFS(config.Templates, "trimAndSummaryPrompt.tmpl")
	tmplSum, _ := template.ParseFS(config.Templates, "summaryOnlyPrompt.tmpl")
	tmplTrim, _ := template.ParseFS(config.Templates, "trimPrompt.tmpl")

	return &workerService{
		bookRepo:   br,
		cacheRepo:  cr,
		actionRepo: ar,
		taskRepo:   tr,
		promptRepo: pr,
		llm:        llm,
		cfg:        cfg,
		tmplBatch:  tmplBatch,
		tmplSum:    tmplSum,
		tmplTrim:   tmplTrim,
	}
}

func (s *workerService) GenerateSummary(ctx context.Context, bookFP string, index int, md5 string, content string) {
	// 获取通用摘要配置 (Type=1)
	summaryPrompt, err := s.promptRepo.GetPromptByID(ctx, 1)
	if err != nil {
		log.Warn().Msg("Summary config (ID=1) not found, using default fallback")
		summaryPrompt = &domain.Prompt{SummaryPromptContent: "请生成本章剧情摘要，200字内。"}
	}

	data := struct {
		SummaryPromptContent string
		Summaries            string
	}{
		SummaryPromptContent: summaryPrompt.SummaryPromptContent,
		Summaries:            "暂无", // 异步补全时暂时不带上下文，或者后续可以拉取
	}

	var buf bytes.Buffer
	if s.tmplSum != nil {
		_ = s.tmplSum.Execute(&buf, data)
	}
	
	summary, err := s.llm.Chat(ctx, buf.String(), content)
	if err != nil {
		return
	}

	err = s.cacheRepo.SaveSummary(ctx, &domain.RawSummary{
		BookFingerprint: bookFP, ChapterIndex: index, Content: summary, CreatedAt: time.Now(),
	})
	if err != nil {
		log.Error().Err(err).Msg("Save Summary Failed")
		return
	}
}

func (s *workerService) StartBatchTrim(ctx context.Context, userID uint, bookID uint, promptID uint) (string, error) {
	taskID := uuid.New().String()
	task := &domain.Task{
		ID: taskID, UserID: userID, BookID: bookID, Type: "batch_trim", Status: "pending", Progress: 0, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := s.taskRepo.CreateTask(ctx, task); err != nil {
		return "", err
	}
	go s.runBatchTrimTask(context.Background(), task, promptID)
	return taskID, nil
}

func (s *workerService) GetTaskStatus(ctx context.Context, taskID string) (*domain.Task, error) {
	return s.taskRepo.GetTaskByID(ctx, taskID)
}

func (s *workerService) runBatchTrimTask(ctx context.Context, task *domain.Task, promptID uint) {
	task.Status = "running"
	if err := s.taskRepo.UpdateTask(ctx, task); err != nil {
		log.Error().Err(err).Str("task_id", task.ID).Msg("Failed to update task status to running")
		return
	}

	book, err := s.bookRepo.GetBookByID(ctx, task.BookID)
	if err != nil {
		log.Error().Err(err).Uint("book_id", task.BookID).Msg("Failed to get book for batch trim")
		task.Status = "failed"
		task.Error = err.Error()
		_ = s.taskRepo.UpdateTask(ctx, task)
		return
	}

	chapters, err := s.bookRepo.GetChaptersByBookID(ctx, task.BookID)
	if err != nil {
		log.Error().Err(err).Uint("book_id", task.BookID).Msg("Failed to get chapters")
		task.Status = "failed"
		task.Error = err.Error()
		_ = s.taskRepo.UpdateTask(ctx, task)
		return
	}

	prompt, err := s.promptRepo.GetPromptByID(ctx, promptID)
	if err != nil {
		log.Error().Err(err).Uint("prompt_id", promptID).Msg("Failed to get prompt")
		task.Status = "failed"
		task.Error = "prompt not found"
		_ = s.taskRepo.UpdateTask(ctx, task)
		return
	}

	// 获取通用摘要配置 (Type=1)
	summaryPrompt, err := s.promptRepo.GetPromptByID(ctx, 1)
	if err != nil {
		log.Warn().Msg("Summary config (ID=1) not found, using default fallback")
		summaryPrompt = &domain.Prompt{SummaryPromptContent: "请生成本章剧情摘要，200字内。"}
	}

	for i, chap := range chapters {
		cache, err := s.cacheRepo.GetTrimResult(ctx, chap.ContentMD5, promptID)
		if err == nil && cache != nil && cache.Level == 2 {
			if err := s.actionRepo.RecordUserTrim(ctx, &domain.UserProcessedChapter{
				UserID: task.UserID, BookID: book.ID, ChapterID: chap.ID, PromptID: promptID, CreatedAt: time.Now(),
			}); err != nil {
				log.Warn().Err(err).Msg("Failed to record user trim during batch")
			}
		} else {
			s.processSingleBatchChapter(ctx, task.UserID, book, chap, prompt, summaryPrompt, 2)
		}

		task.Progress = int(float64(i+1) / float64(len(chapters)) * 100)
		task.UpdatedAt = time.Now()
		if err := s.taskRepo.UpdateTask(ctx, task); err != nil {
			log.Warn().Err(err).Str("task_id", task.ID).Msg("Failed to update task progress")
		}
		if (i+1)%s.cfg.EncyclopediaInterval == 0 {
			s.UpdateEncyclopedia(ctx, book.Fingerprint, i+1)
		}
	}
	task.Status = "completed"
	_ = s.taskRepo.UpdateTask(ctx, task)
}

func (s *workerService) processSingleBatchChapter(ctx context.Context, userID uint, book *domain.Book, chap domain.Chapter, prompt *domain.Prompt, sumPrompt *domain.Prompt, targetLevel int) {
	summaries, err := s.cacheRepo.GetSummaries(ctx, book.Fingerprint, chap.Index, s.cfg.SummaryLimit)
	if err != nil {
		log.Warn().Err(err).Msg("Batch: Failed to get summaries")
	}

	encyclopedia, err := s.cacheRepo.GetEncyclopedia(ctx, book.Fingerprint, chap.Index)
	if err != nil {
		log.Warn().Err(err).Msg("Batch: Failed to get encyclopedia")
	}

	raw, err := s.bookRepo.GetRawContent(ctx, chap.ContentMD5)
	if err != nil {
		log.Error().Err(err).Str("md5", chap.ContentMD5).Msg("Batch: Failed to get raw content")
		return
	}

	// 1. 准备模板数据
	rawLen := len([]rune(raw.Content))
	boundaryRatioMin := prompt.BoundaryRatioMin
	if boundaryRatioMin == 0 {
		boundaryRatioMin = prompt.TargetRatioMin
	}
	boundaryRatioMax := prompt.BoundaryRatioMax
	if boundaryRatioMax == 0 {
		boundaryRatioMax = prompt.TargetRatioMax
	}

	minWords := int(float64(rawLen) * boundaryRatioMin)
	maxWords := int(float64(rawLen) * boundaryRatioMax)

	targetMin := int(prompt.TargetRatioMin * 100)
	targetMax := int(prompt.TargetRatioMax * 100)
	targetRangeStr := fmt.Sprintf("%d-%d", targetMin, targetMax)
	if targetMin == targetMax {
		targetRangeStr = fmt.Sprintf("%d", targetMin)
	}

	data := struct {
		WordsRange              string
		TargetResidualRateRange string
		PromptContent           string
		SummaryPromptContent    string
		Summaries               string
		Encyclopedia            string
	}{
		WordsRange:              fmt.Sprintf("%d-%d", minWords, maxWords),
		TargetResidualRateRange: targetRangeStr,
		PromptContent:           prompt.PromptContent,
		SummaryPromptContent:    sumPrompt.SummaryPromptContent,
		Summaries:               formatSummaries(summaries),
		Encyclopedia:            formatEncyclopedia(encyclopedia),
	}

	// 2. 渲染 Batch 模板
	var buf bytes.Buffer
	if s.tmplBatch == nil {
		log.Error().Msg("Batch: Template not loaded")
		return
	}
	if err := s.tmplBatch.Execute(&buf, data); err != nil {
		log.Error().Err(err).Msg("Batch: Template render failed")
		return
	}
	systemPrompt := buf.String()

	// 3. 调用 LLM
	resText, err := s.llm.Chat(ctx, systemPrompt, raw.Content)
	if err != nil {
		log.Error().Err(err).Int("idx", chap.Index).Msg("Batch Task: LLM Chat failed")
		return
	}

	// 4. 解析 XML
	trimmedContent := extractTagContent(resText, "content")
	summaryContent := extractTagContent(resText, "summary")

	// 5. 容错重试 (Fallback)
	if trimmedContent == "" {
		log.Warn().Int("idx", chap.Index).Msg("Batch: XML parse failed, falling back to separate calls")

		// 5.1 获取正文 (使用 trimPrompt)
		var trimBuf bytes.Buffer
		if s.tmplTrim != nil {
			_ = s.tmplTrim.Execute(&trimBuf, data)
			trimmedContent, err = s.llm.Chat(ctx, trimBuf.String(), raw.Content)
			if err != nil {
				log.Error().Err(err).Msg("Fallback: Trim failed")
				return
			}
		}

		// 5.2 获取摘要 (使用 summaryOnlyPrompt)
		var sumBuf bytes.Buffer
		if s.tmplSum != nil {
			_ = s.tmplSum.Execute(&sumBuf, data)
			summaryContent, _ = s.llm.Chat(ctx, sumBuf.String(), raw.Content)
		}
	}

	if summaryContent == "" {
		summaryContent = "摘要生成失败"
	}

	trimmedLen := len([]rune(trimmedContent))
	rate := 0.0
	if rawLen > 0 {
		rate = float64(trimmedLen) / float64(rawLen)
		rate = float64(int(rate*10000+0.5)) / 10000
	}

	if err := s.cacheRepo.SaveTrimResult(ctx, &domain.TrimResult{
		ContentMD5:     chap.ContentMD5,
		PromptID:       prompt.ID,
		Level:          targetLevel,
		TrimmedContent: trimmedContent,
		TrimWords:      trimmedLen,
		TrimRate:       rate,
		CreatedAt:      time.Now(),
	}); err != nil {
		log.Error().Err(err).Msg("Batch: Failed to save trim result")
	}

	if err := s.cacheRepo.SaveSummary(ctx, &domain.RawSummary{
		BookFingerprint: book.Fingerprint, ChapterIndex: chap.Index, Content: summaryContent, CreatedAt: time.Now(),
	}); err != nil {
		log.Error().Err(err).Msg("Batch: Failed to save summary")
	}

	if err := s.actionRepo.RecordUserTrim(ctx, &domain.UserProcessedChapter{
		UserID: userID, BookID: book.ID, ChapterID: chap.ID, PromptID: prompt.ID, CreatedAt: time.Now(),
	}); err != nil {
		log.Warn().Err(err).Msg("Batch: Failed to record user trim")
	}
}

func (s *workerService) UpdateEncyclopedia(ctx context.Context, bookFP string, endIdx int) {
	old, err := s.cacheRepo.GetEncyclopedia(ctx, bookFP, endIdx)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to fetch old encyclopedia, assuming nil")
	}

	summaries, err := s.cacheRepo.GetSummaries(ctx, bookFP, endIdx+1, s.cfg.EncyclopediaInterval)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch summaries for encyclopedia update")
		return
	}
	if len(summaries) == 0 {
		return
	}
	var summaryTexts []string
	for _, sm := range summaries {
		summaryTexts = append(summaryTexts, sm.Content)
	}
	prompt := "你是一个文学设定分析员。请基于[旧百科]和[最新剧情摘要]，合并更新为最新的Markdown设定集。"
	input := fmt.Sprintf("[旧百科]\n%s\n\n[新剧情摘要]\n%s", func() string {
		if old != nil {
			return old.Content
		}
		return "暂无"
	}(), strings.Join(summaryTexts, "\n"))

	content, err := s.llm.Chat(ctx, prompt, input)
	if err != nil {
		return
	}

	err = s.cacheRepo.SaveEncyclopedia(ctx, &domain.SharedEncyclopedia{
		BookFingerprint: bookFP, RangeEnd: endIdx, Content: content, CreatedAt: time.Now(),
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to save encyclopedia")
	}
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
