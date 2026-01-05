package service

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github/zqr233qr/story-trim/internal/domain"

	"github.com/rs/zerolog/log"
)

type SplitterService struct {
	chapterPattern *regexp.Regexp
}

func NewSplitterService() *SplitterService {
	pattern := regexp.MustCompile(`(?m)^\s*第[0-9一二三四五六七八九十百千零两]+章.*$`)
	return &SplitterService{
		chapterPattern: pattern,
	}
}

func (s *SplitterService) SplitFile(filePath string) ([]domain.SplitChapter, error) {
	log.Debug().Str("file", filePath).Msg("Starting to split file")

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	chapters := s.SplitContent(string(content))
	log.Info().Int("count", len(chapters)).Msg("Chapters split successfully")
	return chapters, nil
}

func (s *SplitterService) SplitContent(content string) []domain.SplitChapter {
	matches := s.chapterPattern.FindAllStringIndex(content, -1)
	var chapters []domain.SplitChapter

	if len(matches) == 0 {
		log.Warn().Msg("No chapters detected, treating whole content as one chapter")
		chapters = append(chapters, domain.SplitChapter{
			Index:   0,
			Title:   "全文",
			Content: content,
		})
		return chapters
	}

	// 序章处理
	if matches[0][0] > 0 {
		preface := content[:matches[0][0]]
		if strings.TrimSpace(preface) != "" {
			chapters = append(chapters, domain.SplitChapter{
				Index:   0,
				Title:   "序章/前言",
				Content: strings.TrimSpace(preface),
			})
		}
	}

	for i, loc := range matches {
		start := loc[0]
		end := len(content)
		if i < len(matches)-1 {
			end = matches[i+1][0]
		}

		title := strings.TrimSpace(content[start:loc[1]])
		body := strings.TrimSpace(content[loc[1]:end])

		chapters = append(chapters, domain.SplitChapter{
			Index:   len(chapters) + 1,
			Title:   title,
			Content: body,
		})
	}

	return chapters
}
