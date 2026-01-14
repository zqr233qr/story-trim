package repository

import (
	"errors"
	"fmt"

	"github.com/zqr233qr/story-trim/internal/config"
	"github.com/zqr233qr/story-trim/internal/model"
	"github.com/zqr233qr/story-trim/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	gormlogger "gorm.io/gorm/logger"
)

func NewDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(cfg.Source), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(
		&model.Book{},
		&model.Chapter{},
		&model.ChapterContent{},
		&model.Prompt{},
		&model.Task{},
		&model.TrimResult{},
		&model.UserProcessedChapter{},
		&model.ReadingHistory{},
		&model.User{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto migrate database: %w", err)
	}

	// 初始化数据库中的提示数据
	err = promptSeeder(db)
	if err != nil {
		return nil, fmt.Errorf("failed to seed prompts: %w", err)
	}

	logger.Info().Msg("Database connected and migrated successfully")
	return db, nil
}

// promptSeeder 初始化数据库中的提示数据
func promptSeeder(db *gorm.DB) error {
	prompts := []model.Prompt{
		{
			ID:             1,
			Name:           "标准沉浸模式",
			Description:    "大幅删减无意义的重复描写、心理独白和环境堆砌。保留核心对话与伏笔。",
			IsSystem:       true,
			IsDefault:      true, // Default
			TargetRatioMin: 0.50, TargetRatioMax: 0.60,
			BoundaryRatioMin: 0.45, BoundaryRatioMax: 0.65,
			PromptContent: `1. **去水去冗**：大幅删减无意义的重复描写、心理独白和环境堆砌。
2. **场景整合**：将冗长的过场段落改写为简练的白描。
3. **保留核心**：全量保留对话，保留推动剧情的关键动作和伏笔细节。`,
		},
		{
			ID:             2,
			Name:           "轻度精简模式",
			Description:    "优化语感、合并琐碎短句。全量保留对话和环境渲染，适合细读。",
			IsSystem:       true,
			IsDefault:      false,
			TargetRatioMin: 0.75, TargetRatioMax: 0.85,
			BoundaryRatioMin: 0.70, BoundaryRatioMax: 0.90,
			PromptContent: `1. **语感修饰**：优化并合并原文中过于琐碎、重复的短句；精减无实际语义的语气助词（如“的、了、吧、呢”的过度堆砌）。
2. **全量保留**：全量保留所有对话内容、环境渲染、角色的独特神态描写以及烘托意境的关键细节。
3. **最小干预**：除非是明显的废话，否则不要删除。`,
		},
		{
			ID:             3,
			Name:           "极简速读模式",
			Description:    "剧情优先。大胆删除所有环境与心理描写，紧凑叙事，快速通关。",
			IsSystem:       true,
			IsDefault:      false,
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

// FirstRecodeIgnoreError 获取第一条记录，忽略错误
func FirstRecodeIgnoreError(db *gorm.DB, dest interface{}) (bool, error) {
	if err := db.First(&dest).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ExistWithoutObject 检查表中是否存在指定的记录，不返回记录本身
func ExistWithoutObject(db *gorm.DB) (bool, error) {
	var count int64
	err := db.Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
