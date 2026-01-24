package parser

import (
	"os"
	"strings"
	"testing"
)

func TestChineseNumeralTitleWithPunctuation(t *testing.T) {
	rules := []Rule{
		{
			Name:    "Merged_Chinese",
			Pattern: `(?m)^(?:第)?[0-9零一二三四五六七八九十百千万]+(?:[章回节][ \t\f]?|[ \t\f]+).*`, // 支持 第一章 / 0001章 / 0001 标题
			Weight:  90,
		},
	}

	var sb strings.Builder
	chapters := []string{
		"第一章 午时，决斗开始！",
		"第二章 午时，决斗开始！",
		"第四十八章 午时，决斗开始！",
		"0001章 仙界归来",
		"0002 仙界归来",
	}
	for _, title := range chapters {
		sb.WriteString(title + "\n")
		sb.WriteString(strings.Repeat("正文...", 120))
		sb.WriteString("\n")
	}

	indices, ruleName := SmartParseTXT(sb.String(), rules)
	if len(indices) != 5 {
		t.Errorf("Expected 5 chapters, got %d", len(indices))
	}
	if ruleName != "Merged_Chinese" {
		t.Errorf("Expected rule Merged_Chinese, got %s", ruleName)
	}
}

func TestTestSmartParseLocalTXTFile(t *testing.T) {
	file, err := os.ReadFile("/Users/zqr/Downloads/sonovel-macos_x64/downloads/仙帝归来(风无极光).txt")
	//file, err := os.ReadFile("/Users/zqr/Downloads/sonovel-macos_x64/downloads/仙逆(耳根).txt")
	if err != nil {
		t.Errorf("Failed to read file, error: %v", err)
		return
	}

	chapterIndices, ruleName := SmartParseTXT(string(file), nil)
	t.Logf("Selected Rule: %s, Chapters: %d", ruleName, len(chapterIndices))
	for _, index := range chapterIndices {
		t.Logf("Index: %d, Title: %s, Start: %d, End: %d, Len: %d", index.Index, index.Title, index.Start, index.End, index.Len)
	}
}

func formatInt(n int) string {
	// 简单转 string，测试用
	return string(rune('0' + n%10)) // 仅个位数测试用
}
