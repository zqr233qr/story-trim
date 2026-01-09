package service

import (
	"fmt"
	"strings"
	"github/zqr233qr/story-trim/internal/core/domain"
)

func formatSummaries(summaries []domain.ChapterSummary) string {
	if len(summaries) == 0 {
		return "无"
	}
	var sb strings.Builder
	for _, s := range summaries {
		sb.WriteString(fmt.Sprintf("第%d章摘要: %s\n", s.ChapterIndex, s.Content))
	}
	return sb.String()
}

func formatEncyclopedia(enc *domain.SharedEncyclopedia) string {
	if enc == nil {
		return "无"
	}
	return enc.Content
}

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
