package integrity

import (
	"context"
	"io"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockClient is a mock implementation of storage.Client
type MockClient struct {
	mock.Mock
}

func (m *MockClient) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	args := m.Called(ctx, bucketName)
	return args.Bool(0), args.Error(1)
}

func (m *MockClient) MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error {
	args := m.Called(ctx, bucketName, opts)
	return args.Error(0)
}

func (m *MockClient) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	args := m.Called(ctx, bucketName, objectName, reader, objectSize, opts)
	return args.Get(0).(minio.UploadInfo), args.Error(1)
}

func (m *MockClient) ListObjects(ctx context.Context, bucketName string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	args := m.Called(ctx, bucketName, opts)
	// For simplicity, return a channel that can be fed by the test setup, or nil/closed one.
	// We can use a type assertion to return what mock returns.
	if ch, ok := args.Get(0).(<-chan minio.ObjectInfo); ok {
		return ch
	}
	// Fallback to empty closed channel if mock returns nil
	ch := make(chan minio.ObjectInfo)
	close(ch)
	return ch
}

func TestCheckStructure(t *testing.T) {
	logger := zap.NewNop()

	t.Run("Bucket Missing", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("BucketExists", mock.Anything, "assets").Return(false, nil)

		svc := NewService(mockClient, "assets", logger)
		_, err := svc.CheckStructure(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("All Missing", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("BucketExists", mock.Anything, "assets").Return(true, nil)
		// ListObjects returns empty channel
		ch := make(chan minio.ObjectInfo)
		close(ch)
		mockClient.On("ListObjects", mock.Anything, "assets", mock.Anything).Return((<-chan minio.ObjectInfo)(ch))

		svc := NewService(mockClient, "assets", logger)
		missing, err := svc.CheckStructure(context.Background())
		assert.NoError(t, err)
		assert.Len(t, missing, len(RequiredFolders))
	})

	t.Run("All Present", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("BucketExists", mock.Anything, "assets").Return(true, nil)
		
		// Map logic: ListObjects called for each folder. Return channel with 1 item.
		// Mock needs to handle multiple calls with different inputs? ListObjects args contain Prefix.
		// Since testify mock arg matching is strict, we can setup matching for each folder.
		
		for _, folder := range RequiredFolders {
			ch := make(chan minio.ObjectInfo, 1)
			ch <- minio.ObjectInfo{Key: folder + "/"}
			close(ch)
			// Match opts with Prefix
			mockClient.On("ListObjects", mock.Anything, "assets", mock.MatchedBy(func(opts minio.ListObjectsOptions) bool {
				return opts.Prefix == folder + "/"
			})).Return((<-chan minio.ObjectInfo)(ch))
		}

		svc := NewService(mockClient, "assets", logger)
		missing, err := svc.CheckStructure(context.Background())
		assert.NoError(t, err)
		assert.Len(t, missing, 0)
	})
}

func TestFixStructure(t *testing.T) {
	logger := zap.NewNop()
	mockClient := new(MockClient)

	// Fix validation
	// PutObject should be called for each missing folder
	mockClient.On("PutObject", mock.Anything, "assets", mock.Anything, mock.Anything, int64(0), mock.Anything).Return(minio.UploadInfo{}, nil)

	svc := NewService(mockClient, "assets", logger)
	err := svc.FixStructure(context.Background(), []string{"bundled"})
	assert.NoError(t, err)
	mockClient.AssertNumberOfCalls(t, "PutObject", 1)
}
