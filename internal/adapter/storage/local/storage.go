package local

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type storage struct {
	baseDir string
}

func NewStorage(baseDir string) (*storage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}
	return &storage{baseDir: baseDir}, nil
}

func (s *storage) Save(ctx context.Context, originalName string, data []byte) (string, error) {
	// 生成唯一文件名防止重名
	ext := filepath.Ext(originalName)
	newName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	path := filepath.Join(s.baseDir, newName)

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}
	return path, nil
}

func (s *storage) Delete(ctx context.Context, path string) error {
	return os.Remove(path)
}
