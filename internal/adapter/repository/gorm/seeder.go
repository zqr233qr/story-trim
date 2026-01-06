package gorm

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SeedPrompts(db *gorm.DB) error {
	prompts := []Prompt{
		// Type 1: Summary Config (Special)
		{
			ID:       1,
			Name:     "Standard Summary Config",
			Type:     1, // 摘要配置
			IsSystem: true,
			SummaryPromptContent: `1. **字数限制**：200 - 400 字。
2. **内容重点**：
   - 核心事件：发生了什么？
   - 人物动态：主要角色做了什么决定？
   - 信息增量：有什么新的设定或伏笔被揭示？
3. **格式**：直接输出摘要内容，不要分点，不要使用 Markdown 标题。`,
		},
		// Type 0: Trim Configs
		{
			ID:       2,
			Name:     "标准沉浸模式",
			Type:     0, // 精简配置
			IsSystem: true,
			TargetRatioMin: 0.50, TargetRatioMax: 0.60,
			BoundaryRatioMin: 0.45, BoundaryRatioMax: 0.65,
			PromptContent: `1. **去水去冗**：大幅删减无意义的重复描写、心理独白和环境堆砌。
2. **场景整合**：将冗长的过场段落改写为简练的白描。
3. **保留核心**：全量保留对话，保留推动剧情的关键动作和伏笔细节。`,
		},
		{
			ID:       3,
			Name:     "轻度精简模式",
			Type:     0,
			IsSystem: true,
			TargetRatioMin: 0.75, TargetRatioMax: 0.85,
			BoundaryRatioMin: 0.70, BoundaryRatioMax: 0.90,
			PromptContent: `1. **语感修饰**：优化并合并原文中过于琐碎、重复的短句；精减无实际语义的语气助词（如“的、了、吧、呢”的过度堆砌）。
2. **全量保留**：全量保留所有对话内容、环境渲染、角色的独特神态描写以及烘托意境的关键细节。
3. **最小干预**：除非是明显的废话，否则不要删除。`,
		},
		{
			ID:       4,
			Name:     "极简速读模式",
			Type:     0,
			IsSystem: true,
			TargetRatioMin: 0.25, TargetRatioMax: 0.35,
			BoundaryRatioMin: 0.20, BoundaryRatioMax: 0.40,
			PromptContent: `1. **剧情优先**：只保留推动剧情发展的核心事件和关键对话。
2. **大胆删除**：所有的环境描写、心理活动、次要人物的寒暄全部删除。
3. **结构重组**：在不破坏时间线的前提下，紧凑叙事节奏。`,
		},
	}

	for _, p := range prompts {
		// Use ID as constraint for seeding
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			UpdateAll: true,
		}).Create(&p).Error; err != nil {
			return err
		}
	}
	return nil
}