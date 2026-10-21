package filestore

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/zserub/machinerysafety-BE/internal/config"
)

// MinioStore provides MinIO storage operations.
type MinioStore struct {
	client *minio.Client
	bucket string
	logger *slog.Logger
}

// NewMinioStore creates a new MinIO store.
func NewMinioStore(cfg *config.Config, logger *slog.Logger) (*MinioStore, error) {
	logger = logger.With("component", "filestore.minio")

	logger.Info("connecting to MinIO",
		"endpoint", cfg.MinIOEndpoint,
		"bucket", cfg.MinIOBucket,
		"use_ssl", cfg.MinIOUseSSL,
	)

	client, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKeyID, cfg.MinIOSecretAccessKey, ""),
		Secure: cfg.MinIOUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	store := &MinioStore{
		client: client,
		bucket: cfg.MinIOBucket,
		logger: logger,
	}

	return store, nil
}

// EnsureBucket creates the bucket if it doesn't exist.
func (s *MinioStore) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		s.logger.Info("creating bucket", "bucket", s.bucket)
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	s.logger.Info("bucket ready", "bucket", s.bucket)
	return nil
}

// Upload stores a file in MinIO.
func (s *MinioStore) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	s.logger.Debug("uploading file",
		"bucket", s.bucket,
		"key", key,
		"size_bytes", size,
		"content_type", contentType,
	)

	start := time.Now()
	info, err := s.client.PutObject(ctx, s.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		s.logger.Error("upload failed", "bucket", s.bucket, "key", key, "error", err)
		return fmt.Errorf("failed to upload file: %w", err)
	}

	s.logger.Debug("upload completed",
		"bucket", s.bucket,
		"key", key,
		"etag", info.ETag,
		"duration_ms", time.Since(start).Milliseconds(),
	)

	return nil
}

// Download retrieves a file from MinIO.
func (s *MinioStore) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	s.logger.Debug("downloading file", "bucket", s.bucket, "key", key)

	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		s.logger.Error("download failed", "bucket", s.bucket, "key", key, "error", err)
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return obj, nil
}

// Delete removes a file from MinIO.
func (s *MinioStore) Delete(ctx context.Context, key string) error {
	s.logger.Debug("deleting file", "bucket", s.bucket, "key", key)

	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		s.logger.Error("delete failed", "bucket", s.bucket, "key", key, "error", err)
		return fmt.Errorf("failed to delete file: %w", err)
	}

	s.logger.Debug("file deleted", "bucket", s.bucket, "key", key)
	return nil
}

// GetURL returns a presigned URL for temporary access.
func (s *MinioStore) GetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucket, key, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return url.String(), nil
}

// Ping verifies MinIO connectivity.
func (s *MinioStore) Ping(ctx context.Context) error {
	_, err := s.client.BucketExists(ctx, s.bucket)
	return err
}
