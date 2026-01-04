package service_test

import (
	"os"
	"testing"

	"github/zqr233qr/story-trim/internal/service"
)

func TestSplitterService_SplitContent_DPCQ(t *testing.T) {
	// 读取项目根目录下的 dpcq.txt
	// 假设运行测试时在 internal/service 目录下，需要向上两级
	// 或者我们直接使用绝对路径或相对路径，这里尝试相对路径
	content, err := os.ReadFile("../../dpcq.txt")
	if err != nil {
		t.Skipf("dpcq.txt not found, skipping integration test: %v", err)
	}

	splitter := service.NewSplitterService()
	chapters := splitter.SplitContent(string(content))

	if len(chapters) == 0 {
		t.Fatal("Expected chapters, got 0")
	}

	t.Logf("Detected %d chapters", len(chapters))

	// 验证第一章
	firstChap := chapters[0]
	t.Logf("Chapter 1 Title: %s", firstChap.Title)
	if firstChap.Title != "第一章 陨落的天才" {
		t.Errorf("Expected title '第一章 陨落的天才', got '%s'", firstChap.Title)
	}

	// 简单的逻辑验证：每章内容不应为空
	for _, chap := range chapters {
		if len(chap.Content) < 10 {
			t.Errorf("Chapter %s content is too short", chap.Title)
		}
	}
}

func TestSplitterService_SplitContent_Regex(t *testing.T) {
	// 单元测试：测试各种奇葩标题格式
	text := `
		序章 这是一个序章
		...
		第1章 标准标题
		内容...
		  第二章  带空格的标题  
		内容...
		第三千零一章 大数字标题
		内容...
	`
	splitter := service.NewSplitterService()
	chapters := splitter.SplitContent(text)

	expectedTitles := []string{"序章/前言", "第1章 标准标题", "第二章  带空格的标题", "第三千零一章 大数字标题"}
	
	if len(chapters) != 4 {
		t.Fatalf("Expected 4 chapters, got %d", len(chapters))
	}

	for i, title := range expectedTitles {
		if chapters[i].Title != title {
			t.Errorf("Index %d: expected '%s', got '%s'", i, title, chapters[i].Title)
		}
	}
}
