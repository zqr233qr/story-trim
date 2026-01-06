package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github/zqr233qr/story-trim/internal/core/domain"
	"github/zqr233qr/story-trim/internal/core/port"
)

type TrimConfig struct {
	SummaryLimit         int
	EncyclopediaInterval int
	BaseInstruction      string
	MockStreamSpeed      int
}

type trimService struct {
	cacheRepo  port.CacheRepository
	bookRepo   port.BookRepository
	actionRepo port.ActionRepository
	promptRepo port.PromptRepository
	workerSvc  port.WorkerService
	llm        port.LLMPort
	cfg        *TrimConfig
}

func NewTrimService(cr port.CacheRepository, br port.BookRepository, ar port.ActionRepository, pr port.PromptRepository, ws port.WorkerService, llm port.LLMPort, cfg *TrimConfig) *trimService {
	return &trimService{
		cacheRepo:  cr,
		bookRepo:   br,
		actionRepo: ar,
		promptRepo: pr,
		workerSvc:  ws,
		llm:        llm,
		cfg:        cfg,
	}
}

func (s *trimService) TrimChapterStream(ctx context.Context, userID uint, chapterID uint, promptID uint) (<-chan string, error) {
	chap, err := s.bookRepo.GetChapterByID(ctx, chapterID)
	if err != nil { return nil, err }
	book, err := s.bookRepo.GetBookByID(ctx, chap.BookID)
	if err != nil { return nil, err }
	prompt, err := s.promptRepo.GetPromptByID(ctx, promptID)
	if err != nil { return nil, err }

	summaries, _ := s.cacheRepo.GetSummaries(ctx, book.Fingerprint, chap.Index, s.cfg.SummaryLimit)
	encyclopedia, _ := s.cacheRepo.GetEncyclopedia(ctx, book.Fingerprint, chap.Index)

	level := s.calculateLevel(len(summaries), encyclopedia != nil)

	cache, err := s.cacheRepo.GetTrimResult(ctx, chap.ContentMD5, promptID, prompt.Version)
	if err == nil && cache != nil && cache.Level >= level {
		_ = s.actionRepo.RecordUserTrim(ctx, &domain.UserProcessedChapter{
			UserID: userID, BookID: book.ID, ChapterID: chap.ID, PromptID: promptID, CreatedAt: time.Now(),
		})
		return s.mockStreaming(cache.TrimmedContent), nil
	}

	raw, err := s.bookRepo.GetRawContent(ctx, chap.ContentMD5)
	if err != nil { return nil, err }

	// 组装带动态数值的提示词
	systemPrompt := s.buildSystemPrompt(encyclopedia, summaries, prompt, raw.Content)
	
	llmStream, err := s.llm.ChatStream(ctx, systemPrompt, raw.Content)
	if err != nil { return nil, err }

	return s.wrapAndSaveStream(ctx, llmStream, userID, book, chap, promptID, prompt.Version, level), nil
}

func (s *trimService) calculateLevel(summaryCount int, hasEncyclopedia bool) int {
	diffCount := summaryCount < s.cfg.SummaryLimit
	if summaryCount == 0 && !hasEncyclopedia { return 0 }
	if diffCount || hasEncyclopedia {
		if !diffCount { return 2 }
		return 1
	}
	return 0
}

func (s *trimService) buildSystemPrompt(enc *domain.SharedEncyclopedia, summaries []domain.RawSummary, prompt *domain.Prompt, rawContent string) string {
	var sb strings.Builder
	
	// 1. 系统底层协议 (V4.5 - 对话保护版)
	sb.WriteString("### 核心编辑协议 (必须严格遵守):\n")
	sb.WriteString("- 身份：你是一名拥有顶级文学素养的小说主编。目标是优化阅读节奏，提升文学美感。\n")
	sb.WriteString("- 对话红线：严禁对原文中的对话进行改写、删减、合并或人称切换。必须全量、原样保留所有对话字句（除非是极度重复的无效口头禅）。\n")
	sb.WriteString("- 禁令：严禁剧透。输出内容严禁引入原文未提及的后续背景。严禁使用总结性语言。输出必须是自然的小说正文。\n")
	sb.WriteString("- 优先级：对话完整性 > 逻辑连贯性 > 字数压缩目标。\n\n")

	// 2. 动态计算字数目标
	rawLen := len([]rune(rawContent))
	var minRate, maxRate float64
	switch prompt.ID {
	case 1: // 轻度
		minRate, maxRate = 0.70, 0.85
	case 3: // 极简
		minRate, maxRate = 0.15, 0.35
	default: // 标准及其他
		minRate, maxRate = 0.45, 0.65
	}

	minWords := int(float64(rawLen) * minRate)
	maxWords := int(float64(rawLen) * maxRate)

	// 3. 注入模板 (替换占位符)
	template := strings.ReplaceAll(prompt.Content, "{MIN_WORDS}", fmt.Sprintf("%d", minWords))
	template = strings.ReplaceAll(template, "{MAX_WORDS}", fmt.Sprintf("%d", maxWords))

	sb.WriteString("### [特定精简要求]\n")
	sb.WriteString(template)
	sb.WriteString("\n\n")

	if enc != nil {
		sb.WriteString("### [逻辑参考：全局百科]\n")
		sb.WriteString(enc.Content + "\n\n")
	}
	
	if len(summaries) > 0 {
		sb.WriteString("### [逻辑参考：前情提要]\n")
		for _, sm := range summaries { sb.WriteString("- " + sm.Content + "\n") }
		sb.WriteString("\n")
	}

	final := sb.String()
	log.Debug().Str("prompt", final).Msg("Assembled System Prompt")
	return final
}

func (s *trimService) mockStreaming(content string) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		runes := []rune(content)
		for i := 0; i < len(runes); {
			step := 10
			if i+step > len(runes) { step = len(runes) - i }
			ch <- string(runes[i:i+step])
			i += step
			time.Sleep(time.Duration(s.cfg.MockStreamSpeed) * time.Millisecond)
		}
	}()
	return ch
}

func (s *trimService) wrapAndSaveStream(ctx context.Context, input <-chan string, userID uint, book *domain.Book, chap *domain.Chapter, pID uint, pVer string, level int) <-chan string {
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
			raw, _ := s.bookRepo.GetRawContent(ctx, chap.ContentMD5)
			rawLen := len([]rune(raw.Content))
			trimmedLen := len([]rune(final))
			
			rate := 0.0
			if rawLen > 0 {
				rate = float64(trimmedLen) / float64(rawLen)
				rate = float64(int(rate*10000+0.5)) / 10000
			}

			log.Info().
				Str("md5", chap.ContentMD5).
				Int("trimmed_words", trimmedLen).
				Float64("trim_rate", rate).
				Msg("Streaming trim completed")

			_ = s.cacheRepo.SaveTrimResult(ctx, &domain.TrimResult{
				ContentMD5:     chap.ContentMD5,
				PromptID:       pID,
				PromptVersion:  pVer,
				Level:          level,
				TrimmedContent: final,
				TrimWords:      trimmedLen,
				TrimRate:       rate,
				CreatedAt:      time.Now(),
			})
			if userID > 0 {
				_ = s.actionRepo.RecordUserTrim(ctx, &domain.UserProcessedChapter{
					UserID: userID, BookID: book.ID, ChapterID: chap.ID, PromptID: pID, CreatedAt: time.Now(),
				})
			}
			go s.workerSvc.GenerateSummary(context.Background(), book.Fingerprint, chap.Index, chap.ContentMD5, final)
		}
	}()
	return output
}
