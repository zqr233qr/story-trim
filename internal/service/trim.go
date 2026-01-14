package service

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/zqr233qr/story-trim/internal/errno"
	"github.com/zqr233qr/story-trim/internal/model"
	"github.com/zqr233qr/story-trim/internal/repository"
	"github.com/zqr233qr/story-trim/pkg/logger"
	"github.com/zqr233qr/story-trim/templates"
)

type TrimService struct {
	bookRepo   repository.BookRepositoryInterface
	tmpl       *template.Template
	llmService LlmServiceInterface
}

func NewTrimService(bookRepo repository.BookRepositoryInterface, llmService LlmServiceInterface) *TrimService {
	tmpl, err := template.ParseFS(templates.FS, "trimPrompt.tmpl")
	if err != nil {
		panic("failed to load templates: " + err.Error())
	}
	return &TrimService{bookRepo: bookRepo, tmpl: tmpl, llmService: llmService}
}

func (s *TrimService) RenderPrompt(name string, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := s.tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (s *TrimService) TrimStreamByMD5(ctx context.Context, userID uint, chapterMD5 string, rawContent string, promptID uint) (<-chan string, error) {
	cache, err := s.bookRepo.GetTrimResult(ctx, chapterMD5, promptID)
	if err == nil && cache != nil {
		if userID > 0 {
			go s.bookRepo.RecordUserTrim(context.Background(), &model.UserProcessedChapter{
				UserID:     userID,
				PromptID:   promptID,
				ChapterMD5: chapterMD5,
				CreatedAt:  time.Now(),
			})
		}
		return s.mockStreaming(cache.TrimContent), nil
	}

	return s.mockStreaming(rawContent[:min(100, len(rawContent))]), nil
}

func (s *TrimService) TrimStreamByChapterID(ctx context.Context, userID uint, bookID uint, chapterID uint, promptID uint) (<-chan string, error) {
	chap, err := s.bookRepo.GetChapterByID(ctx, chapterID)
	if err != nil {
		return nil, errno.ErrChapterNotFound
	}

	cache, err := s.bookRepo.GetTrimResult(ctx, chap.ChapterMD5, promptID)
	if err == nil && cache != nil {
		if userID > 0 {
			go s.bookRepo.RecordUserTrim(context.Background(), &model.UserProcessedChapter{
				UserID:     userID,
				BookID:     bookID,
				ChapterID:  chapterID,
				PromptID:   promptID,
				ChapterMD5: chap.ChapterMD5,
				CreatedAt:  time.Now(),
			})
		}
		return s.mockStreaming(cache.TrimContent), nil
	}

	raw, err := s.bookRepo.GetRawContent(ctx, chap.ChapterMD5)
	if err != nil {
		return nil, errno.ErrChapterNotFound
	}

	prompt, err := s.bookRepo.GetPromptByID(ctx, promptID)
	if err != nil {
		return nil, err
	}

	systemPrompt := s.buildSystemPrompt(prompt, raw.Content)

	llmResp, err := s.llmService.LlmWithStream(ctx, systemPrompt.systemPrompt, raw.Content)
	if err != nil {
		return nil, err
	}

	// 创建流式响应 channel
	ch := make(chan string)
	go func() {
		defer close(ch)

		var fullContent strings.Builder

		t := time.Now()
		// 流式读取 LLM 响应
		for {
			resp, err := llmResp.Stream.Recv()
			if err != nil {
				// 流结束或出错
				break
			}

			if len(resp.Choices) > 0 {
				content := resp.Choices[0].Delta.Content
				if content != "" {
					fullContent.WriteString(content)
					ch <- content
				}
			}

			// 获取 token 使用情况（在流的最后一条消息中）
			if resp.Usage != nil {
				llmResp.TotalTokens = resp.Usage.TotalTokens
				llmResp.PromptTokens = resp.Usage.PromptTokens
				llmResp.CompletionTokens = resp.Usage.CompletionTokens

				llmResp.InputCost = float64(llmResp.PromptTokens) * llmResp.InputMTokenPrice / million
				llmResp.OutputCost = float64(llmResp.CompletionTokens) * llmResp.OutputMTokenPrice / million
				llmResp.TotalCost = llmResp.InputCost + llmResp.OutputCost
			}
		}

		trimmedContent := fullContent.String()
		if trimmedContent != "" {
			// 计算字数和压缩率
			trimWords := len([]rune(trimmedContent))
			rawWords := len([]rune(raw.Content))
			trimRate := ((float64(trimWords)/float64(rawWords))*10000 + 0.5) / 100.0
			takeTime := time.Since(t).Seconds()

			// 保存处理结果到缓存
			trimResult := &model.TrimResult{
				ChapterMD5:       chap.ChapterMD5,
				PromptID:         promptID,
				TrimContent:      trimmedContent,
				TrimContentWords: trimWords,
				WordsRange:       systemPrompt.WordsRange,
				TrimRate:         trimRate,
				TargetRateRange:  systemPrompt.TargetRateRange,
				TotalCost:        llmResp.TotalCost,
				InputCost:        llmResp.InputCost,
				OutputCost:       llmResp.OutputCost,
				TotalTokens:      llmResp.TotalTokens,
				PromptTokens:     llmResp.PromptTokens,
				CompletionTokens: llmResp.CompletionTokens,
				TakeTime:         takeTime,
				LlmName:          llmResp.LlmName,
				CreatedAt:        time.Time{},
			}

			if err := s.bookRepo.SaveTrimResult(context.Background(), trimResult); err != nil {
				logger.Error().Err(err).Msg("failed to save trim result")
				return
			}
		}

		// 记录用户处理记录
		if userID > 0 {
			if err := s.bookRepo.RecordUserTrim(context.Background(), &model.UserProcessedChapter{
				UserID:     userID,
				BookID:     bookID,
				ChapterID:  chapterID,
				PromptID:   promptID,
				ChapterMD5: chap.ChapterMD5,
			}); err != nil {
				logger.Error().Err(err).Msg("failed to record user trim")
				return
			}
		}
	}()

	return ch, nil
}

type systemPromptData struct {
	systemPrompt    string
	WordsRange      string
	TargetRateRange string
}

func (s *TrimService) buildSystemPrompt(prompt *model.Prompt, rawContent string) systemPromptData {
	rawLen := len([]rune(rawContent))
	minWords := int(float64(rawLen) * prompt.TargetRatioMin)
	maxWords := int(float64(rawLen) * prompt.TargetRatioMax)

	data := struct {
		WordsRange      string
		TargetRateRange string
		PromptContent   string
		Summaries       string
		Encyclopedia    string
	}{
		WordsRange:      fmt.Sprintf("%d-%d", minWords, maxWords),
		TargetRateRange: fmt.Sprintf("%d-%d%", int(prompt.TargetRatioMin*100), int(prompt.TargetRatioMax*100)),
		PromptContent:   prompt.PromptContent,
	}

	var buf bytes.Buffer
	_ = s.tmpl.Execute(&buf, data)
	return systemPromptData{
		systemPrompt:    buf.String(),
		WordsRange:      data.WordsRange,
		TargetRateRange: data.TargetRateRange,
	}
}

func (s *TrimService) mockStreaming(content string) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		for i := 0; i < len(content); i += 10 {
			end := i + 10
			if end > len(content) {
				end = len(content)
			}
			ch <- content[i:end]
		}
	}()
	return ch
}

func (s *TrimService) TrimChatByChapterID(ctx context.Context, userID uint, chapterID uint, promptID uint) error {
	chapter, err := s.bookRepo.GetChapterByID(ctx, chapterID)
	if err != nil {
		return err
	}
	if chapter == nil {
		return errno.ErrChapterNotFound
	}

	exist, err := s.bookRepo.ExistTrimResultWithoutObject(ctx, chapter.ChapterMD5, promptID)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}

	prompt, err := s.bookRepo.GetPromptByID(ctx, promptID)
	if err != nil {
		return err
	}

	rawContent, err := s.bookRepo.GetRawContent(ctx, chapter.ChapterMD5)
	if err != nil {
		return err
	}

	systemPrompt := s.buildSystemPrompt(prompt, rawContent.Content)

	t := time.Now()
	llmResp, err := s.llmService.Llm(ctx, systemPrompt.systemPrompt, rawContent.Content)
	if err != nil {
		return err
	}

	takeTime := time.Since(t)

	trimContent := llmResp.Resp.Choices[0].Message.Content
	trimContentWords := len([]rune(trimContent))
	rawContentWords := len([]rune(rawContent.Content))
	// 保留两位小数 百分比
	trimRate := ((float64(trimContentWords)/float64(rawContentWords))*10000 + 0.5) / 100.0

	trimResult := &model.TrimResult{
		ChapterMD5:       chapter.ChapterMD5,
		PromptID:         promptID,
		TrimContent:      trimContent,
		TrimContentWords: trimContentWords,
		WordsRange:       systemPrompt.WordsRange,
		TrimRate:         trimRate,
		TargetRateRange:  systemPrompt.TargetRateRange,
		TotalCost:        llmResp.TotalCost,
		InputCost:        llmResp.InputCost,
		OutputCost:       llmResp.OutputCost,
		TotalTokens:      llmResp.TotalTokens,
		PromptTokens:     llmResp.PromptTokens,
		CompletionTokens: llmResp.CompletionTokens,
		TakeTime:         takeTime.Seconds(),
		LlmName:          llmResp.LlmName,
	}

	if err := s.bookRepo.SaveTrimResult(context.Background(), trimResult); err != nil {
		logger.Error().Err(err).Msg("failed to save trim result")
		return err
	}

	// 记录用户处理记录
	if err := s.bookRepo.RecordUserTrim(context.Background(), &model.UserProcessedChapter{
		UserID:     userID,
		BookID:     chapter.BookID,
		ChapterID:  chapterID,
		PromptID:   promptID,
		ChapterMD5: chapter.ChapterMD5,
	}); err != nil {
		logger.Error().Err(err).Msg("failed to record user trim")
		return err
	}

	return nil
}

type TrimServiceInterface interface {
	TrimStreamByMD5(ctx context.Context, userID uint, chapterMD5 string, rawContent string, promptID uint) (<-chan string, error)
	TrimStreamByChapterID(ctx context.Context, userID uint, bookID uint, chapterID uint, promptID uint) (<-chan string, error)
	TrimChatByChapterID(ctx context.Context, userID uint, chapterID uint, promptID uint) error
}
