package checks

import (
	"context"
	"testing"

	"asset-manager/core/storage/mocks"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestCheckStructure(t *testing.T) {
	t.Run("Bucket Missing", func(t *testing.T) {
		mockClient := new(mocks.Client)
		mockClient.On("BucketExists", mock.Anything, "assets").Return(false, nil)

		_, err := CheckStructure(context.Background(), mockClient, "assets")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("All Missing", func(t *testing.T) {
		mockClient := new(mocks.Client)
		mockClient.On("BucketExists", mock.Anything, "assets").Return(true, nil)
		ch := make(chan minio.ObjectInfo)
		close(ch)
		mockClient.On("ListObjects", mock.Anything, "assets", mock.Anything).Return((<-chan minio.ObjectInfo)(ch))

		missing, err := CheckStructure(context.Background(), mockClient, "assets")
		assert.NoError(t, err)
		assert.Len(t, missing, len(RequiredFolders))
	})

	t.Run("All Present", func(t *testing.T) {
		mockClient := new(mocks.Client)
		mockClient.On("BucketExists", mock.Anything, "assets").Return(true, nil)

		for _, folder := range RequiredFolders {
			ch := make(chan minio.ObjectInfo, 1)
			ch <- minio.ObjectInfo{Key: folder + "/"}
			close(ch)
			mockClient.On("ListObjects", mock.Anything, "assets", mock.MatchedBy(func(opts minio.ListObjectsOptions) bool {
				return opts.Prefix == folder+"/"
			})).Return((<-chan minio.ObjectInfo)(ch))
		}

		missing, err := CheckStructure(context.Background(), mockClient, "assets")
		assert.NoError(t, err)
		assert.Len(t, missing, 0)
	})
	t.Run("Bucket Check Error", func(t *testing.T) {
		mockClient := new(mocks.Client)
		mockClient.On("BucketExists", mock.Anything, "assets").Return(false, assert.AnError)

		_, err := CheckStructure(context.Background(), mockClient, "assets")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check bucket existence")
	})
}

func TestFixStructure(t *testing.T) {
	logger := zap.NewNop()

	t.Run("Success", func(t *testing.T) {
		mockClient := new(mocks.Client)
		mockClient.On("PutObject", mock.Anything, "assets", mock.Anything, mock.Anything, int64(0), mock.Anything).Return(minio.UploadInfo{}, nil)

		err := FixStructure(context.Background(), mockClient, "assets", logger, []string{"bundled"})
		assert.NoError(t, err)
		mockClient.AssertNumberOfCalls(t, "PutObject", 1)
	})

	t.Run("PutObject Error", func(t *testing.T) {
		mockClient := new(mocks.Client)
		mockClient.On("PutObject", mock.Anything, "assets", mock.Anything, mock.Anything, int64(0), mock.Anything).Return(minio.UploadInfo{}, assert.AnError)

		err := FixStructure(context.Background(), mockClient, "assets", logger, []string{"bundled"})
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})
}
