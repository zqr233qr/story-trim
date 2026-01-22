package storage

import (
	"context"
	"io"
)

// Storage 定义统一的对象存储操作接口。
type Storage interface {
	// Put 写入对象内容。
	Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	// Get 读取对象内容流。
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	// Exists 判断对象是否存在。
	Exists(ctx context.Context, key string) (bool, error)
	// Delete 删除对象。
	Delete(ctx context.Context, key string) error
}
