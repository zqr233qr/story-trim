package database

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github/zqr233qr/story-trim/internal/domain"
	"github/zqr233qr/story-trim/pkg/config"
)

// Init 初始化数据库连接并执行迁移
func Init(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "sqlite":
		dialector = sqlite.Open(cfg.Source)
	// case "mysql":
	// 	dialector = mysql.Open(cfg.Source)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(&domain.User{}, &domain.Book{}, &domain.Chapter{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}