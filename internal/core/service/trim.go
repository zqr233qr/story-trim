package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github/zqr233qr/story-trim/internal/core/domain"
	"github/zqr233qr/story-trim/internal/core/port"
	"github/zqr233qr/story-trim/pkg/config"

	"github.com/rs/zerolog/log"
)

type TrimConfig struct {
	SummaryLimit         int
	EncyclopediaInterval int
	MockStreamSpeed      int
	PromptRates          map[string][]float64
}

type trimService struct {
	cacheRepo  port.CacheRepository
	bookRepo   port.BookRepository
	actionRepo port.ActionRepository
	promptRepo port.PromptRepository
	workerSvc  port.WorkerService
	llm        port.LLMPort
	cfg        *TrimConfig
	tmpl       *template.Template
}

func NewTrimService(cr port.CacheRepository, br port.BookRepository, ar port.ActionRepository, pr port.PromptRepository, ws port.WorkerService, llm port.LLMPort, cfg *TrimConfig) *trimService {
	// 从嵌入的文件系统中加载模板
	tmpl, err := template.ParseFS(config.Templates, "trimPrompt.tmpl")
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse embedded trimPrompt.tmpl")
	}
	
	return &trimService{
		cacheRepo:  cr,
		bookRepo:   br,
		actionRepo: ar,
		promptRepo: pr,
		workerSvc:  ws,
		llm:        llm,
		cfg:        cfg,
		tmpl:       tmpl,
	}
}

func (s *trimService) TrimChapterStream(ctx context.Context, userID uint, chapterID uint, promptID uint) (<-chan string, error) {
	chap, err := s.bookRepo.GetChapterByID(ctx, chapterID)
	if err != nil {
		return nil, err
	}
	book, err := s.bookRepo.GetBookByID(ctx, chap.BookID)
	if err != nil {
		return nil, err
	}
	prompt, err := s.promptRepo.GetPromptByID(ctx, promptID)
	if err != nil {
		return nil, err
	}

	summaries, err := s.cacheRepo.GetSummaries(ctx, book.Fingerprint, chap.Index, s.cfg.SummaryLimit)
	if err != nil {
		log.Warn().Err(err).Uint("book_id", book.ID).Int("chap_idx", chap.Index).Msg("Failed to fetch summaries, proceeding without them")
	}

	encyclopedia, err := s.cacheRepo.GetEncyclopedia(ctx, book.Fingerprint, chap.Index)
	if err != nil {
		log.Warn().Err(err).Uint("book_id", book.ID).Int("chap_idx", chap.Index).Msg("Failed to fetch encyclopedia, proceeding without it")
	}

	level := s.calculateLevel(len(summaries), encyclopedia != nil)

	cache, err := s.cacheRepo.GetTrimResult(ctx, chap.ContentMD5, promptID)
	if err == nil && cache != nil && cache.Level >= level {
		if err := s.actionRepo.RecordUserTrim(ctx, &domain.UserProcessedChapter{
			UserID: userID, BookID: book.ID, ChapterID: chap.ID, PromptID: promptID, CreatedAt: time.Now(),
		}); err != nil {
			log.Warn().Err(err).Uint("user_id", userID).Uint("chap_id", chap.ID).Msg("Failed to record user trim action")
		}
		return s.mockStreaming(cache.TrimmedContent), nil
	}

	raw, err := s.bookRepo.GetRawContent(ctx, chap.ContentMD5)
	if err != nil {
		return nil, err
	}

	systemPrompt := s.buildSystemPrompt(encyclopedia, summaries, prompt, raw.Content)

	llmStream, err := s.llm.ChatStream(ctx, systemPrompt, raw.Content)
	if err != nil {
		return nil, err
	}

	return s.wrapAndSaveStream(ctx, llmStream, userID, book, chap, chap.ContentMD5, promptID, level), nil
}

// TrimContentStream 无状态精简：直接根据内容哈希进行处理
func (s *trimService) TrimContentStream(ctx context.Context, userID uint, rawContent string, promptID uint) (<-chan string, error) {
	prompt, err := s.promptRepo.GetPromptByID(ctx, promptID)
	if err != nil {
		return nil, err
	}

	md5 := s.calculateMD5(rawContent)

	// 2. 查缓存
	cache, err := s.cacheRepo.GetTrimResult(ctx, md5, promptID)
	if err == nil && cache != nil {
		return s.mockStreaming(cache.TrimmedContent), nil
	}

	// 3. 确保原文入库 (RawContent)
	go func() {
		// s.bookRepo.SaveRawContent(...) // 需要 repo 支持
	}()

	// 4. 组装 Prompt
	systemPrompt := s.buildSystemPrompt(nil, nil, prompt, rawContent)

	llmStream, err := s.llm.ChatStream(ctx, systemPrompt, rawContent)
	if err != nil {
		return nil, err
	}

	return s.wrapAndSaveStream(ctx, llmStream, userID, nil, nil, md5, promptID, 0), nil
}

func (s *trimService) calculateMD5(content string) string {
	hash := md5.Sum([]byte(content))
	return hex.EncodeToString(hash[:])
}

func (s *trimService) calculateLevel(summaryCount int, hasEncyclopedia bool) int {
	diffCount := summaryCount < s.cfg.SummaryLimit
	if summaryCount == 0 && !hasEncyclopedia {
		return 0
	}
	if diffCount || hasEncyclopedia {
		if !diffCount {
			return 2
		}
		return 1
	}
	return 0
}

func (s *trimService) buildSystemPrompt(enc *domain.SharedEncyclopedia, summaries []domain.RawSummary, prompt *domain.Prompt, rawContent string) string {
	rawLen := len([]rune(rawContent))

	boundaryRatioMin := prompt.BoundaryRatioMin
	if boundaryRatioMin == 0 { boundaryRatioMin = prompt.TargetRatioMin }
	boundaryRatioMax := prompt.BoundaryRatioMax
	if boundaryRatioMax == 0 { boundaryRatioMax = prompt.TargetRatioMax }

	minWords := int(float64(rawLen) * boundaryRatioMin)
	maxWords := int(float64(rawLen) * boundaryRatioMax)
	
	targetMin := int(prompt.TargetRatioMin * 100)
	targetMax := int(prompt.TargetRatioMax * 100)
	targetRangeStr := fmt.Sprintf("%d-%d", targetMin, targetMax)
	if targetMin == targetMax {
		targetRangeStr = fmt.Sprintf("%d", targetMin)
	}

	data := struct {
		WordsRange             string
		TargetResidualRateRange string
		PromptContent          string
		Summaries              string
		Encyclopedia           string
	}{
		WordsRange:              fmt.Sprintf("%d-%d", minWords, maxWords),
		TargetResidualRateRange: targetRangeStr,
		PromptContent:           prompt.PromptContent,
		Summaries:               formatSummaries(summaries),
		Encyclopedia:            formatEncyclopedia(enc),
	}

	if s.tmpl == nil {
		return "System Error: Template not loaded"
	}
	var buf bytes.Buffer
	if err := s.tmpl.Execute(&buf, data); err != nil {
		log.Error().Err(err).Msg("Failed to render prompt template")
		return ""
	}

	final := buf.String()
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
			if i+step > len(runes) {
				step = len(runes) - i
			}
			ch <- string(runes[i : i+step])
			i += step
			time.Sleep(time.Duration(s.cfg.MockStreamSpeed) * time.Millisecond)
		}
	}()
	return ch
}

func (s *trimService) wrapAndSaveStream(ctx context.Context, input <-chan string, userID uint, book *domain.Book, chap *domain.Chapter, contentMD5 string, pID uint, level int) <-chan string {
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
			// 如果是无状态模式，我们无法查 RawContent 表获取原文长度，只能假设
			// 或者，我们在 TrimContentStream 里把原文也传进来？不需要，直接用 contentMD5 存即可
			// 为了计算 rate，我们可能需要原文长度。
			// 简单起见，如果 chap 为 nil，我们就不算 rate 了，或者默认 0
			
			trimmedLen := len([]rune(final))
			rate := 0.0
			
			// 如果 chap 不为空，去查原文算 rate
			if chap != nil {
				raw, _ := s.bookRepo.GetRawContent(ctx, contentMD5)
				if raw != nil {
					rawLen := len([]rune(raw.Content))
					if rawLen > 0 {
						rate = float64(trimmedLen) / float64(rawLen)
						rate = float64(int(rate*10000+0.5)) / 10000
					}
				}
			}

			log.Info().
				Str("md5", contentMD5).
				Int("trimmed_words", trimmedLen).
				Float64("trim_rate", rate).
				Msg("Streaming trim completed")

			if err := s.cacheRepo.SaveTrimResult(ctx, &domain.TrimResult{
				ContentMD5:     contentMD5,
				PromptID:       pID,
				Level:          level,
				TrimmedContent: final,
				TrimWords:      trimmedLen,
				TrimRate:       rate,
				CreatedAt:      time.Now(),
			}); err != nil {
				log.Error().Err(err).Str("md5", contentMD5).Msg("Failed to save trim result")
			}

			if userID > 0 && book != nil && chap != nil {
				if err := s.actionRepo.RecordUserTrim(ctx, &domain.UserProcessedChapter{
					UserID: userID, BookID: book.ID, ChapterID: chap.ID, PromptID: pID, CreatedAt: time.Now(),
				}); err != nil {
					log.Warn().Err(err).Uint("user_id", userID).Msg("Failed to record user trim action")
				}
			}
            
            if book != nil && chap != nil {
			    go s.workerSvc.GenerateSummary(context.Background(), book.Fingerprint, chap.Index, contentMD5, final)
            }
		}
	}()
	return output
}