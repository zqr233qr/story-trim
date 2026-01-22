package storage

import (
	"fmt"
	"strings"

	"github.com/zqr233qr/story-trim/internal/config"
)

// NewStorage 根据配置创建存储实现。
func NewStorage(cfg config.StorageConfig) (Storage, error) {
	switch strings.ToLower(cfg.Type) {
	case "minio":
		return NewMinIOStorage(cfg.MinIO)
	case "":
		return nil, fmt.Errorf("storage type is required")
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.Type)
	}
}
