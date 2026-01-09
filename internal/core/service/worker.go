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
	queue      *TaskQueue
	tmplBatch  *template.Template
	tmplSum    *template.Template
	tmplTrim   *template.Template
}

func NewWorkerService(br port.BookRepository, cr port.CacheRepository, ar port.ActionRepository, tr port.TaskRepository, pr port.PromptRepository, llm port.LLMPort, cfg *TrimConfig, queue *TaskQueue) *workerService {
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
		queue:      queue,
		tmplBatch:  tmplBatch,
		tmplSum:    tmplSum,
		tmplTrim:   tmplTrim,
	}
}

func (s *workerService) SubmitFullTrimTask(ctx context.Context, userID uint, bookID uint, promptID uint) (string, error) {
	if _, err := s.bookRepo.GetBookByID(ctx, bookID); err != nil {
		return "", fmt.Errorf("book not found: %w", err)
	}

	taskID := uuid.New().String()
	task := &domain.Task{
		ID: taskID, UserID: userID, BookID: bookID, Type: "full_trim", Status: "pending", Progress: 0, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := s.taskRepo.CreateTask(ctx, task); err != nil {
		return "", err
	}

	s.queue.Submit(&FullTrimJob{
		s:        s,
		task:     task,
		promptID: promptID,
	})

	return taskID, nil
}

func (s *workerService) GenerateSummary(ctx context.Context, bookFP string, index int, md5 string, content string) {
	s.queue.Submit(&SummaryJob{
		s:       s,
		bookFP:  bookFP,
		index:   index,
		md5:     md5,
		content: content,
	})
}

func (s *workerService) UpdateEncyclopedia(ctx context.Context, bookFP string, endIdx int) {
	s.queue.Submit(&EncyclopediaJob{
		s:      s,
		bookFP: bookFP,
		endIdx: endIdx,
	})
}

type FullTrimJob struct {
	s        *workerService
	task     *domain.Task
	promptID uint
}

func (j *FullTrimJob) Execute(ctx context.Context) error {
	j.task.Status = "running"
	_ = j.s.taskRepo.UpdateTask(ctx, j.task)

	book, _ := j.s.bookRepo.GetBookByID(ctx, j.task.BookID)
	chapters, _ := j.s.bookRepo.GetChaptersByBookID(ctx, j.task.BookID)
	prompt, _ := j.s.promptRepo.GetPromptByID(ctx, j.promptID)
	summaryPrompt, _ := j.s.promptRepo.GetSummaryPrompt(ctx)

	for i, chap := range chapters {
		cache, err := j.s.cacheRepo.GetTrimResult(ctx, chap.ChapterMD5, j.promptID)
		if err == nil && cache != nil && cache.Level >= 1 {
			_ = j.s.actionRepo.RecordUserTrim(ctx, &domain.UserProcessedChapter{
				UserID: j.task.UserID, BookID: book.ID, ChapterID: chap.ID, PromptID: j.promptID, ChapterMD5: chap.ChapterMD5, CreatedAt: time.Now(),
			})
		} else {
			j.s.processSingleBatchChapter(ctx, j.task.UserID, book, chap, prompt, summaryPrompt, 1)
		}

		j.task.Progress = int(float64(i+1) / float64(len(chapters)) * 100)
		j.task.UpdatedAt = time.Now()
		_ = j.s.taskRepo.UpdateTask(ctx, j.task)

		if (i+1)%j.s.cfg.EncyclopediaInterval == 0 {
			j.s.UpdateEncyclopedia(ctx, book.Fingerprint, i+1)
		}
	}
	j.task.Status = "completed"
	return j.s.taskRepo.UpdateTask(ctx, j.task)
}

type SummaryJob struct {
	s       *workerService
	bookFP  string
	index   int
	md5     string
	content string
}

func (j *SummaryJob) Execute(ctx context.Context) error {
	summaryPrompt, _ := j.s.promptRepo.GetSummaryPrompt(ctx)
	if summaryPrompt == nil {
		summaryPrompt = &domain.Prompt{SummaryPromptContent: "请生成本章剧情摘要，200字内。"}
	}

	data := struct{ SummaryPromptContent string }{SummaryPromptContent: summaryPrompt.SummaryPromptContent}
	var buf bytes.Buffer
	_ = j.s.tmplSum.Execute(&buf, data)

	summary, usage, err := j.s.llm.Chat(ctx, buf.String(), j.content)
	if err == nil {
		return j.s.cacheRepo.SaveSummary(ctx, &domain.ChapterSummary{
			ChapterMD5:      j.md5,
			BookFingerprint: j.bookFP,
			ChapterIndex:    j.index,
			Content:         summary,
			ConsumeToken:    usage.TotalTokens,
			CreatedAt:       time.Now(),
		})
	}
	return err
}

type EncyclopediaJob struct {
	s      *workerService
	bookFP string
	endIdx int
}

func (j *EncyclopediaJob) Execute(ctx context.Context) error {
	existing, _ := j.s.cacheRepo.GetEncyclopedia(ctx, j.bookFP, j.endIdx)
	if existing != nil && existing.RangeEnd >= j.endIdx {
		return nil
	}

	summaries, _ := j.s.cacheRepo.GetSummaries(ctx, j.bookFP, j.endIdx+1, j.s.cfg.EncyclopediaInterval)
	if len(summaries) == 0 {
		return nil
	}

	var sb strings.Builder
	for _, sm := range summaries {
		sb.WriteString(sm.Content + "\n")
	}

	prompt := "你是一个文学设定分析员。请基于[旧百科]和[最新剧情摘要]，合并更新为最新的Markdown设定集。"
	input := fmt.Sprintf("[旧百科]\n%s\n\n[新摘要]\n%s", func() string { if existing != nil { return existing.Content }; return "无" }(), sb.String())

	content, usage, err := j.s.llm.Chat(ctx, prompt, input)
	if err == nil {
		log.Debug().Int("token", usage.TotalTokens).Msg("Encyclopedia Updated")
		return j.s.cacheRepo.SaveEncyclopedia(ctx, &domain.SharedEncyclopedia{
			BookFingerprint: j.bookFP, RangeEnd: j.endIdx, Content: content, CreatedAt: time.Now(),
		})
	}
	return err
}

func (s *workerService) processSingleBatchChapter(ctx context.Context, userID uint, book *domain.Book, chap domain.Chapter, prompt *domain.Prompt, sumPrompt *domain.Prompt, targetLevel int) {
	summaries, _ := s.cacheRepo.GetSummaries(ctx, book.Fingerprint, chap.Index, s.cfg.SummaryLimit)
	encyclopedia, _ := s.cacheRepo.GetEncyclopedia(ctx, book.Fingerprint, chap.Index)
	raw, _ := s.bookRepo.GetRawContent(ctx, chap.ChapterMD5)
	if raw == nil {
		return
	}

	rawLen := len([]rune(raw.Content))
	data := struct {
		WordsRange              string
		TargetResidualRateRange string
		PromptContent           string
		SummaryPromptContent    string
		Summaries               string
		Encyclopedia            string
	}{
		WordsRange:              fmt.Sprintf("%d-%d", int(float64(rawLen)*prompt.TargetRatioMin), int(float64(rawLen)*prompt.TargetRatioMax)),
		TargetResidualRateRange: fmt.Sprintf("%d-%d%%", int(prompt.TargetRatioMin*100), int(prompt.TargetRatioMax*100)),
		PromptContent:           prompt.PromptContent,
		SummaryPromptContent:    sumPrompt.SummaryPromptContent,
		Summaries:               formatSummaries(summaries),
		Encyclopedia:            formatEncyclopedia(encyclopedia),
	}

	var trimmedContent, summaryContent string
	var totalTokens int

	var buf bytes.Buffer
	if err := s.tmplBatch.Execute(&buf, data); err == nil {
		resText, usage, err := s.llm.Chat(ctx, buf.String(), raw.Content)
		if err == nil {
			trimmedContent = extractTagContent(resText, "content")
			summaryContent = extractTagContent(resText, "summary")
			totalTokens = usage.TotalTokens
		}
	}

	if trimmedContent == "" {
		var tBuf bytes.Buffer
		_ = s.tmplTrim.Execute(&tBuf, data)
		tText, tUsage, err := s.llm.Chat(ctx, tBuf.String(), raw.Content)
		if err == nil {
			trimmedContent = tText
			totalTokens += tUsage.TotalTokens
		}

		var sBuf bytes.Buffer
		_ = s.tmplSum.Execute(&sBuf, data)
		sText, sUsage, err := s.llm.Chat(ctx, sBuf.String(), raw.Content)
		if err == nil {
			summaryContent = sText
			totalTokens += sUsage.TotalTokens
		}
	}

	if trimmedContent == "" {
		return
	}

	trimmedLen := len([]rune(trimmedContent))
	rate := 0.0
	if rawLen > 0 {
		rate = float64(trimmedLen) / float64(rawLen)
		rate = float64(int(rate*10000+0.5)) / 100.0
	}

	_ = s.cacheRepo.SaveTrimResult(ctx, &domain.TrimResult{
		ChapterMD5:     chap.ChapterMD5,
		PromptID:       prompt.ID,
		Level:          targetLevel,
		TrimmedContent: trimmedContent,
		TrimWords:      trimmedLen,
		TrimRate:       rate,
		ConsumeToken:   totalTokens,
		CreatedAt:      time.Now(),
	})

	if summaryContent != "" {
		_ = s.cacheRepo.SaveSummary(ctx, &domain.ChapterSummary{
			ChapterMD5:      chap.ChapterMD5,
			BookFingerprint: book.Fingerprint,
			ChapterIndex:    chap.Index,
			Content:         summaryContent,
			ConsumeToken:    0,
			CreatedAt:       time.Now(),
		})
	}

	_ = s.actionRepo.RecordUserTrim(ctx, &domain.UserProcessedChapter{
		UserID:     userID,
		BookID:     book.ID,
		ChapterID:  chap.ID,
		PromptID:   prompt.ID,
		ChapterMD5: chap.ChapterMD5,
		CreatedAt:  time.Now(),
	})
}