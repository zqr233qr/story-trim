package gorm

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(path string) (*gorm.DB, error) {
	// 配置 GORM 日志为 Silent，彻底关闭 SQL 打印
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&User{},
		&Book{},
		&Chapter{},
		&ChapterContent{},
		&TrimResult{},
		&ChapterSummary{},
		&SharedEncyclopedia{},
		&UserProcessedChapter{},
		&ReadingHistory{},
		&Task{},
		&Prompt{},
	)
	if err != nil {
		return nil, err
	}

	if err := SeedPrompts(db); err != nil {
		return nil, err
	}

	return db, nil
}
