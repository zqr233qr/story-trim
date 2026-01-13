package repository

import (
	"fmt"

	"github.com/zqr233qr/story-trim/internal/config"
	"github.com/zqr233qr/story-trim/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func NewDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(cfg.Source), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}
	logger.Info().Msg("Database connected successfully")
	return db, nil
}

// FirstRecodeIgnoreError 获取第一条记录，忽略错误
func FirstRecodeIgnoreError(db *gorm.DB, dest interface{}) error {
	return db.Limit(1).Find(dest).Error
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
