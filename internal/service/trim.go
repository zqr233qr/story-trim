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
	bookRepo      repository.BookRepositoryInterface
	pointsService PointsServiceInterface
	tmpl          *template.Template
	llmService    LlmServiceInterface
}

// NewTrimService 创建精简服务。
func NewTrimService(bookRepo repository.BookRepositoryInterface, pointsService PointsServiceInterface, llmService LlmServiceInterface) *TrimService {
	tmpl, err := template.ParseFS(templates.FS, "trimPrompt.tmpl")
	if err != nil {
		panic("failed to load templates: " + err.Error())
	}
	return &TrimService{bookRepo: bookRepo, pointsService: pointsService, tmpl: tmpl, llmService: llmService}
}

func (s *TrimService) RenderPrompt(name string, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := s.tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (s *TrimService) TrimStreamByMD5(ctx context.Context, userID uint, chapterMD5 string, bookMD5 string, bookTitle string, chapterTitle string, rawContent string, promptID uint) (<-chan string, error) {
	if userID > 0 {
		extra := map[string]string{}
		prompt, err := s.bookRepo.GetPromptByID(ctx, promptID)
		if err == nil && prompt != nil {
			extra["prompt_name"] = prompt.Name
		}
		extra["book_title"] = bookTitle
		extra["chapter_title"] = chapterTitle
		extra["book_md5"] = bookMD5
		extra["chapter_md5"] = chapterMD5
		charged, err := s.ensureTrimPoints(ctx, userID, promptID, 0, bookMD5, chapterMD5, "chapter_md5", chapterMD5, extra)
		if err != nil {
			return nil, err
		}
		if charged {
			logger.Info().Msg("points charged for md5 trim")
		}
	}

	cache, err := s.bookRepo.GetTrimResult(ctx, chapterMD5, promptID)
	if err == nil && cache != nil {
		if userID > 0 {
			go s.bookRepo.RecordUserTrim(context.Background(), &model.UserProcessedChapter{
				UserID:     userID,
				PromptID:   promptID,
				BookMD5:    bookMD5,
				ChapterMD5: chapterMD5,
				CreatedAt:  time.Now(),
			})
		}
		return s.mockStreaming(cache.TrimContent), nil
	}

	return s.trimChapter(ctx, userID, 0, 0, bookMD5, chapterMD5, rawContent, promptID)
}

func (s *TrimService) TrimStreamByChapterID(ctx context.Context, userID uint, bookID uint, chapterID uint, promptID uint) (<-chan string, error) {
	chap, err := s.bookRepo.GetChapterByID(ctx, chapterID)
	if err != nil {
		return nil, errno.ErrChapterNotFound
	}

	book, err := s.bookRepo.GetBookByID(ctx, bookID)
	if err != nil {
		return nil, err
	}
	if book == nil {
		return nil, errno.ErrBookNotFound
	}
	bookMD5 := book.BookMD5

	if userID > 0 {
		extra := map[string]string{}
		extra["book_title"] = book.Title
		extra["chapter_title"] = chap.Title
		if prompt, err := s.bookRepo.GetPromptByID(ctx, promptID); err == nil && prompt != nil {
			extra["prompt_name"] = prompt.Name
		}
		extra["book_md5"] = bookMD5
		extra["chapter_md5"] = chap.ChapterMD5
		charged, err := s.ensureTrimPoints(ctx, userID, promptID, bookID, bookMD5, chap.ChapterMD5, "chapter_id", fmt.Sprintf("%d", chapterID), extra)
		if err != nil {
			return nil, err
		}
		if charged {
			logger.Info().Msg("points charged for chapter trim")
		}
	}

	cache, err := s.bookRepo.GetTrimResult(ctx, chap.ChapterMD5, promptID)
	if err == nil && cache != nil {
		if userID > 0 {
			go s.bookRepo.RecordUserTrim(context.Background(), &model.UserProcessedChapter{
				UserID:     userID,
				BookID:     bookID,
				ChapterID:  chapterID,
				PromptID:   promptID,
				BookMD5:    bookMD5,
				ChapterMD5: chap.ChapterMD5,
				CreatedAt:  time.Now(),
			})
		}
		return s.mockStreaming(cache.TrimContent), nil
	}

	return s.trimChapter(ctx, userID, bookID, chapterID, bookMD5, chap.ChapterMD5, "", promptID)
}

func (s *TrimService) trimChapter(ctx context.Context, userID uint, bookID uint, chapterID uint, bookMD5 string, chapterMD5 string, rawContent string, promptID uint) (<-chan string, error) {
	content := ""

	if chapterID > 0 {
		raw, err := s.bookRepo.GetRawContent(ctx, chapterMD5)
		if err != nil {
			return nil, errno.ErrChapterNotFound
		}
		content = raw.Content
	} else {
		content = rawContent
	}

	prompt, err := s.bookRepo.GetPromptByID(ctx, promptID)
	if err != nil {
		return nil, err
	}

	systemPrompt := s.buildSystemPrompt(prompt, content)

	llmResp, err := s.llmService.LlmWithStream(ctx, systemPrompt.systemPrompt, content)
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
			rawWords := len([]rune(content))
			trimRate := ((float64(trimWords)/float64(rawWords))*10000 + 0.5) / 100.0
			takeTime := time.Since(t).Seconds()

			// 保存处理结果到缓存
			trimResult := &model.TrimResult{
				ChapterMD5:       chapterMD5,
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
				BookMD5:    bookMD5,
				ChapterMD5: chapterMD5,
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
		ModeName        string
		WordsRange      string
		TargetRateRange string
		PromptContent   string
	}{
		ModeName:        prompt.Name,
		WordsRange:      fmt.Sprintf("%d-%d", minWords, maxWords),
		TargetRateRange: fmt.Sprintf("%d-%d%%", int(prompt.TargetRatioMin*100), int(prompt.TargetRatioMax*100)),
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
	contentRune := []rune(content)
	go func() {
		defer close(ch)
		for i := 0; i < len([]rune(content)); i += 10 {
			end := i + 10
			if end > len(contentRune) {
				end = len(contentRune)
			}
			ch <- string(contentRune[i:end])
			time.Sleep(time.Millisecond * 100)
		}
	}()
	return ch
}

// ensureTrimPoints 校验并扣除积分，返回是否扣费。
func (s *TrimService) ensureTrimPoints(ctx context.Context, userID uint, promptID uint, bookID uint, bookMD5 string, chapterMD5 string, refType, refID string, extra map[string]string) (bool, error) {
	handled, err := s.bookRepo.HasUserProcessedChapter(ctx, userID, promptID, bookID, bookMD5, chapterMD5)
	if err != nil {
		return false, err
	}
	if handled {
		return false, nil
	}
	if err := s.pointsService.SpendForTrim(ctx, userID, 1, refType, refID, extra); err != nil {
		return false, err
	}
	return true, nil
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
	book, err := s.bookRepo.GetBookByID(ctx, chapter.BookID)
	if err != nil {
		return err
	}
	bookMD5 := ""
	if book != nil {
		bookMD5 = book.BookMD5
	}
	if err := s.bookRepo.RecordUserTrim(context.Background(), &model.UserProcessedChapter{
		UserID:     userID,
		BookID:     chapter.BookID,
		ChapterID:  chapterID,
		PromptID:   promptID,
		BookMD5:    bookMD5,
		ChapterMD5: chapter.ChapterMD5,
	}); err != nil {
		logger.Error().Err(err).Msg("failed to record user trim")
		return err
	}

	return nil
}

type TrimServiceInterface interface {
	TrimStreamByMD5(ctx context.Context, userID uint, chapterMD5 string, bookMD5 string, bookTitle string, chapterTitle string, rawContent string, promptID uint) (<-chan string, error)
	TrimStreamByChapterID(ctx context.Context, userID uint, bookID uint, chapterID uint, promptID uint) (<-chan string, error)
	TrimChatByChapterID(ctx context.Context, userID uint, chapterID uint, promptID uint) error
}
