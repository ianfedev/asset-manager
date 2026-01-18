package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client defines the interface for storage operations.
type Client interface {
	// BucketExists checks if a bucket exists.
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	// MakeBucket creates a new bucket.
	MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error
	// PutObject uploads an object.
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	// GetObject downloads an object.
	GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error)
	// ListObjects lists objects in a bucket.
	ListObjects(ctx context.Context, bucketName string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo
	// RemoveObject removes an object from storage.
	RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error
	// RemoveObjects removes multiple objects from storage.
	RemoveObjects(ctx context.Context, bucketName string, objectsCh <-chan minio.ObjectInfo, opts minio.RemoveObjectsOptions) <-chan minio.RemoveObjectError
}

// NewClient creates a new Minio client based on the configuration.
func NewClient(cfg Config) (Client, error) {
	// Minio expects endpoint without scheme
	endpoint := strings.TrimPrefix(cfg.Endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}
	return &minioClientWrapper{Client: minioClient}, nil
}

type minioClientWrapper struct {
	*minio.Client
}

func (c *minioClientWrapper) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error) {
	return c.Client.GetObject(ctx, bucketName, objectName, opts)
}

func (c *minioClientWrapper) RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error {
	return c.Client.RemoveObject(ctx, bucketName, objectName, opts)
}

func (c *minioClientWrapper) RemoveObjects(ctx context.Context, bucketName string, objectsCh <-chan minio.ObjectInfo, opts minio.RemoveObjectsOptions) <-chan minio.RemoveObjectError {
	return c.Client.RemoveObjects(ctx, bucketName, objectsCh, opts)
}
