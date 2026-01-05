package database

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github/zqr233qr/story-trim/internal/domain"
)

// SeedPrompts 初始化系统预设提示词
func SeedPrompts(db *gorm.DB) error {
	prompts := []domain.Prompt{
		{
			ID:       1,
			Name:     "轻度去水 (10-20%)",
			Version:  "v1.0",
			IsSystem: true,
			Content:  "你是一个资深的网文校对。请在不删减任何剧情、对话和必要环境描写的前提下，去除文本中的错别字、极端冗余的口头禅以及作者重复骗字数的病句。保持原有语境不动。目标：字数减少约 10-20%。",
		},
		{
			ID:       2,
			Name:     "标准精简 (30-40%)",
			Version:  "v1.0",
			IsSystem: true,
			Content:  "你是一个专业的文学编辑。请保留所有对话、关键动作和剧情转折。删除冗余的环境描写、重复的心理活动和无关的填充文字。保持原有的叙事风格。目标：字数减少约 30-40%。",
		},
		{
			ID:       3,
			Name:     "极简模式 (50-60%)",
			Version:  "v1.0",
			IsSystem: true,
			Content:  "你是一个快节奏网文主编。请仅保留主线剧情、关键角色冲突和核心对话。将所有的环境铺垫和次要心理描写压缩到最简，直接展示剧情推进。目标：字数减少 50-60% 以上。",
		},
	}

	for _, p := range prompts {
		p.CreatedAt = time.Now()
		p.UpdatedAt = time.Now()
		// 使用 Upsert 逻辑，如果 ID 存在则更新内容
		err := db.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&p).Error
		if err != nil {
			return err
		}
	}
	return nil
}
