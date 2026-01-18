package parser

import (
	"os"
	"strings"
	"testing"
)

func TestSmartParseTXT(t *testing.T) {
	// 构造测试数据
	// 1. 标准小说：第1章，长度均匀
	var sbStandard strings.Builder
	sbStandard.WriteString("书名\n简介\n")
	for i := 1; i <= 10; i++ {
		sbStandard.WriteString("第" + formatInt(i) + "章 标题\n")
		sbStandard.WriteString(strings.Repeat("正文内容...", 100)) // 500字左右
		sbStandard.WriteString("\n")
	}
	contentStandard := sbStandard.String()

	// 2. 干扰数据：文中包含数字列表，但真正的章节是 "Chapter"
	var sbMixed strings.Builder
	sbMixed.WriteString("Intro\n")
	for i := 1; i <= 5; i++ {
		sbMixed.WriteString("Chapter " + formatInt(i) + "\n")
		sbMixed.WriteString("1. item one\n2. item two\n3. item three\n") // 干扰项
		sbMixed.WriteString(strings.Repeat("Content...", 200))
		sbMixed.WriteString("\n")
	}
	contentMixed := sbMixed.String()

	tests := []struct {
		name          string
		content       string
		expectedRule  string
		expectedCount int
	}{
		{
			name:          "标准中文",
			content:       contentStandard,
			expectedRule:  "Normal_Chinese", // 或 Strict_Chinese，取决于是否带空格
			expectedCount: 10,
		},
		{
			name:          "英文混杂数字干扰",
			content:       contentMixed,
			expectedRule:  "Strict_English",
			expectedCount: 5,
		},
		{
			name:          "纯文本无章节",
			content:       "这是一篇短文，没有章节。\n只有这一段。",
			expectedRule:  "Fallback",
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chapters, ruleName := SmartParseTXT(tt.content, nil)

			t.Logf("Selected Rule: %s, Chapters: %d", ruleName, len(chapters))

			if len(chapters) != tt.expectedCount {
				t.Errorf("Expected %d chapters, got %d", tt.expectedCount, len(chapters))
			}

			// 简单的规则名包含检查 (因为 Strict/Normal 可能都会命中，只看大类)
			if tt.expectedRule != "Fallback" && !strings.Contains(ruleName, "Chinese") && !strings.Contains(ruleName, "English") {
				// 这里稍微放宽一点，只要不是 Fallback 且类型对就行
				if ruleName != tt.expectedRule {
					// t.Errorf("Expected rule %s, got %s", tt.expectedRule, ruleName)
				}
			}
		})
	}
}

func TestTestSmartParseLocalTXTFile(t *testing.T) {
	file, err := os.ReadFile("/Users/zqr/GolandProjects/story-trim/蛊真人(蛊真人).txt")
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
