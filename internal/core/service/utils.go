package service

import (
	"strings"
	"github/zqr233qr/story-trim/internal/core/domain"
)

func formatSummaries(summaries []domain.RawSummary) string {
	if len(summaries) == 0 {
		return "暂无"
	}
	var sb strings.Builder
	for _, s := range summaries {
		sb.WriteString("- " + s.Content + "\n")
	}
	return sb.String()
}

func formatEncyclopedia(enc *domain.SharedEncyclopedia) string {
	if enc == nil {
		return "暂无"
	}
	return enc.Content
}
