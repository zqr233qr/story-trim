package service_test

import (
	"os"
	"strings"
	"testing"

	"github/zqr233qr/story-trim/internal/service"
	"github/zqr233qr/story-trim/pkg/config"
	"github/zqr233qr/story-trim/pkg/logger"
)

func TestLLMService_TrimIntegration(t *testing.T) {
	cfg, err := config.LoadConfig("../../config.yaml")
	if err != nil {
		t.Skipf("config.yaml not found: %v", err)
	}

	logger.Init(cfg.Log.Level, cfg.Log.Format)
	llmSvc := service.NewLLMService(cfg.LLM)

	testContent := "“斗之力，三段！”望着测验魔石碑上面闪亮得甚至有些刺眼的五个大字。"

	t.Log("Testing single snippet...")
	trimmed, err := llmSvc.TrimContent(testContent)
	if err != nil {
		t.Fatalf("LLM call failed: %v", err)
	}
	t.Logf("Trimmed Result: %s", trimmed)
}

func TestLLMService_TrimFullDPCQ(t *testing.T) {
	cfg, err := config.LoadConfig("../../config.yaml")
	if err != nil {
		t.Skipf("config.yaml not found: %v", err)
	}

	logger.Init(cfg.Log.Level, cfg.Log.Format)
	llmSvc := service.NewLLMService(cfg.LLM)
	splitter := service.NewSplitterService()

	content, err := os.ReadFile("../../dpcq.txt")
	if err != nil {
		t.Fatalf("Failed to read dpcq.txt: %v", err)
	}

	chapters := splitter.SplitContent(string(content))
	t.Logf("Total chapters: %d", len(chapters))

	var output strings.Builder
	for _, chap := range chapters {
		t.Logf("Processing: %s", chap.Title)
		
		res, err := llmSvc.TrimContent(chap.Content)
		if err != nil {
			t.Errorf("Error trimming %s: %v", chap.Title, err)
			continue
		}

		output.WriteString("### " + chap.Title + "\n\n")
		output.WriteString(res + "\n\n")
	}

	_ = os.WriteFile("../../dpcq_trimmed_full.txt", []byte(output.String()), 0644)
	t.Log("Done. Check dpcq_trimmed_full.txt")
}