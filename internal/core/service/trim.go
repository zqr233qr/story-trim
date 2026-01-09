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

	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"
)

type TrimConfig struct {
	SummaryLimit         int
	EncyclopediaInterval int
	MockStreamSpeed      int
	PromptRates          map[string][]float64
}

type trimService struct {
	cacheRepo   port.CacheRepository
	bookRepo    port.BookRepository
	actionRepo  port.ActionRepository
	promptRepo  port.PromptRepository
	workerSvc   port.WorkerService
	llm         port.LLMPort
	cfg         *TrimConfig
	tmpl        *template.Template
	summaryTmpl *template.Template
}

func NewTrimService(cr port.CacheRepository, br port.BookRepository, ar port.ActionRepository, pr port.PromptRepository, ws port.WorkerService, llm port.LLMPort, cfg *TrimConfig) *trimService {
	tmpl, err := template.ParseFS(config.Templates, "trimPrompt.tmpl")
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse embedded trimPrompt.tmpl")
	}

	summaryTmpl, err := template.ParseFS(config.Templates, "summaryOnlyPrompt.tmpl")
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse embedded summaryOnlyPrompt.tmpl")
	}

	return &trimService{
		cacheRepo:   cr,
		bookRepo:    br,
		actionRepo:  ar,
		promptRepo:  pr,
		workerSvc:   ws,
		llm:         llm,
		cfg:         cfg,
		tmpl:        tmpl,
		summaryTmpl: summaryTmpl,
	}
}

// TrimStreamByMD5 基于内容寻址的精简 (App/离线优先模式)
func (s *trimService) TrimStreamByMD5(ctx context.Context, userID uint, chapterMD5 string, rawContent string, promptID uint, chapterIndex int, bookFingerprint string) (<-chan string, error) {
	prompt, err := s.promptRepo.GetPromptByID(ctx, promptID)
	if err != nil {
		return nil, err
	}

	// 1. 查缓存
	cache, err := s.cacheRepo.GetTrimResult(ctx, chapterMD5, promptID)
	if err == nil && cache != nil {
		if userID > 0 {
			go s.actionRepo.RecordUserTrim(context.Background(), &domain.UserProcessedChapter{
				UserID:     userID,
				PromptID:   promptID,
				ChapterMD5: chapterMD5,
				CreatedAt:  time.Now(),
			})
		}
		return s.mockStreaming(cache.TrimmedContent), nil
	}

	// 2. 摘要存续检测
	s.ensureSummaryExist(ctx, chapterMD5, bookFingerprint, chapterIndex, rawContent)

	// 3. LLM 精简
	systemPrompt := s.buildSystemPrompt(nil, nil, prompt, rawContent)
	llmStream, usage, err := s.llm.ChatStream(ctx, systemPrompt, rawContent)
	if err != nil {
		return nil, err
	}

	return s.wrapAndSaveStream(ctx, llmStream, usage, userID, nil, nil, chapterMD5, promptID, 0), nil
}

// TrimStreamByChapterID 基于标识寻址的精简 (小程序/云端模式)
func (s *trimService) TrimStreamByChapterID(ctx context.Context, userID uint, bookID uint, chapterID uint, promptID uint) (<-chan string, error) {
	book, err := s.bookRepo.GetBookByID(ctx, bookID)
	if err != nil {
		return nil, err
	}
	chap, err := s.bookRepo.GetChapterByID(ctx, chapterID)
	if err != nil {
		return nil, err
	}
	prompt, err := s.promptRepo.GetPromptByID(ctx, promptID)
	if err != nil {
		return nil, err
	}
	raw, err := s.bookRepo.GetRawContent(ctx, chap.ChapterMD5)
	if err != nil {
		return nil, err
	}

	cache, err := s.cacheRepo.GetTrimResult(ctx, chap.ChapterMD5, promptID)
	if err == nil && cache != nil {
		if userID > 0 {
			go s.actionRepo.RecordUserTrim(context.Background(), &domain.UserProcessedChapter{
				UserID:     userID,
				BookID:     bookID,
				ChapterID:  chapterID,
				PromptID:   promptID,
				ChapterMD5: chap.ChapterMD5,
				CreatedAt:  time.Now(),
			})
		}
		return s.mockStreaming(cache.TrimmedContent), nil
	}

	s.ensureSummaryExist(ctx, chap.ChapterMD5, book.Fingerprint, chap.Index, raw.Content)

	systemPrompt := s.buildSystemPrompt(nil, nil, prompt, raw.Content)
	llmStream, usage, err := s.llm.ChatStream(ctx, systemPrompt, raw.Content)
	if err != nil {
		return nil, err
	}

	return s.wrapAndSaveStream(ctx, llmStream, usage, userID, book, chap, chap.ChapterMD5, promptID, 0), nil
}

func (s *trimService) ensureSummaryExist(ctx context.Context, md5 string, fp string, idx int, rawContent string) {
	isExistSummary, err := s.cacheRepo.IsExistSummary(ctx, md5)
	if err != nil || isExistSummary {
		return
	}

	summaryPrompt, err := s.promptRepo.GetSummaryPrompt(ctx)
	if err == nil && summaryPrompt != nil {
		summarySystemPrompt := s.buildSummarySystemPrompt(summaryPrompt)
		summary, summaryUsage, err := s.llm.Chat(ctx, summarySystemPrompt, rawContent)
		if err == nil {
			_ = s.cacheRepo.SaveSummary(ctx, &domain.ChapterSummary{
				ChapterMD5:      md5,
				BookFingerprint: fp,
				ChapterIndex:    idx,
				Content:         summary,
				ConsumeToken:    summaryUsage.TotalTokens,
				CreatedAt:       time.Now(),
			})
		}
	}
}

func (s *trimService) buildSystemPrompt(enc *domain.SharedEncyclopedia, summaries []domain.ChapterSummary, prompt *domain.Prompt, rawContent string) string {
	rawLen := len([]rune(rawContent))
	minWords := int(float64(rawLen) * prompt.TargetRatioMin)
	maxWords := int(float64(rawLen) * prompt.TargetRatioMax)

	data := struct {
		WordsRange              string
		TargetResidualRateRange string
		PromptContent           string
		Summaries               string
		Encyclopedia            string
	}{
		WordsRange:              fmt.Sprintf("%d-%d", minWords, maxWords),
		TargetResidualRateRange: fmt.Sprintf("%d-%d", int(prompt.TargetRatioMin*100), int(prompt.TargetRatioMax*100)),
		PromptContent:           prompt.PromptContent,
		Summaries:               formatSummaries(summaries),
		Encyclopedia:            formatEncyclopedia(enc),
	}

	var buf bytes.Buffer
	_ = s.tmpl.Execute(&buf, data)
	return buf.String()
}

func (s *trimService) buildSummarySystemPrompt(prompt *domain.Prompt) string {
	data := struct {
		SummaryPromptContent string
	}{
		SummaryPromptContent: prompt.SummaryPromptContent,
	}

	var buf bytes.Buffer
	_ = s.summaryTmpl.Execute(&buf, data)
	return buf.String()
}

func (s *trimService) mockStreaming(content string) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		runes := []rune(content)
		for i := 0; i < len(runes); {
			step := 10
			if i+step > len(runes) { step = len(runes) - i }
			ch <- string(runes[i : i+step])
			i += step
			time.Sleep(time.Duration(s.cfg.MockStreamSpeed) * time.Millisecond)
		}
	}()
	return ch
}

func (s *trimService) wrapAndSaveStream(ctx context.Context, input <-chan string, usage *openai.Usage, userID uint, book *domain.Book, chap *domain.Chapter, chapterMD5 string, pID uint, level int) <-chan string {
	output := make(chan string)
	go func() {
		defer close(output)
		var full strings.Builder
		for text := range input {
			full.WriteString(text)
			output <- text
		}

		final := full.String()
		if final != "" {
			trimmedLen := len([]rune(final))
			rate := 0.0

			raw, _ := s.bookRepo.GetRawContent(ctx, chapterMD5)
			if raw != nil && raw.WordsCount > 0 {
				rate = float64(trimmedLen) / float64(raw.WordsCount)
				rate = float64(int(rate*10000+0.5)) / 100.0
			}

			_ = s.cacheRepo.SaveTrimResult(ctx, &domain.TrimResult{
				ChapterMD5:     chapterMD5,
				PromptID:       pID,
				Level:          level,
				TrimmedContent: final,
				TrimWords:      trimmedLen,
				TrimRate:       rate,
				ConsumeToken:   usage.TotalTokens,
				CreatedAt:      time.Now(),
			})

			if userID > 0 {
				_ = s.actionRepo.RecordUserTrim(context.Background(), &domain.UserProcessedChapter{
					UserID:     userID,
					BookID:     func() uint { if book != nil { return book.ID }; return 0 }(),
					ChapterID:  func() uint { if chap != nil { return chap.ID }; return 0 }(),
					PromptID:   pID,
					ChapterMD5: chapterMD5,
					CreatedAt:  time.Now(),
				})
			}
		}
	}()
	return output
}