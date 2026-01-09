package gorm

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(path string) (*gorm.DB, error) {
	// 开启 SQL 日志以便调试
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
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
