package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/zqr233qr/story-trim/internal/config"
)

// MinIOStorage 提供基于 MinIO 的对象存储实现。
type MinIOStorage struct {
	client *minio.Client
	bucket string
}

// NewMinIOStorage 根据配置创建 MinIO 存储实例。
func NewMinIOStorage(cfg config.MinIOConfig) (*MinIOStorage, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("minio endpoint is required")
	}
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("minio access_key/secret_key is required")
	}
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("minio bucket is required")
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, err
	}

	storage := &MinIOStorage{client: client, bucket: cfg.Bucket}
	if cfg.AutoCreateBucket {
		exists, err := client.BucketExists(context.Background(), cfg.Bucket)
		if err != nil {
			return nil, err
		}
		if !exists {
			if err := client.MakeBucket(context.Background(), cfg.Bucket, minio.MakeBucketOptions{Region: cfg.Region}); err != nil {
				return nil, err
			}
		}
	}

	return storage, nil
}

// Put 写入对象内容。
func (s *MinIOStorage) Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	if key == "" {
		return fmt.Errorf("object key is required")
	}
	if size < 0 {
		return fmt.Errorf("object size must be non-negative")
	}
	if contentType == "" {
		contentType = "text/plain; charset=utf-8"
	}
	_, err := s.client.PutObject(ctx, s.bucket, key, reader, size, minio.PutObjectOptions{ContentType: contentType})
	return err
}

// Get 读取对象内容流。
func (s *MinIOStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	if key == "" {
		return nil, fmt.Errorf("object key is required")
	}
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	if _, err := obj.Stat(); err != nil {
		_ = obj.Close()
		return nil, err
	}
	return obj, nil
}

// Exists 判断对象是否存在。
func (s *MinIOStorage) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("object key is required")
	}
	_, err := s.client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Delete 删除对象。
func (s *MinIOStorage) Delete(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("object key is required")
	}
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}
