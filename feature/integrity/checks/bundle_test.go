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

func TestCheckBundled(t *testing.T) {
	t.Run("Bundled All Missing", func(t *testing.T) {
		mockClient := new(mocks.Client)
		mockClient.On("BucketExists", mock.Anything, "assets").Return(true, nil)
		ch := make(chan minio.ObjectInfo)
		close(ch)
		mockClient.On("ListObjects", mock.Anything, "assets", mock.Anything).Return((<-chan minio.ObjectInfo)(ch))

		missing, err := CheckBundled(context.Background(), mockClient, "assets")
		assert.NoError(t, err)
		assert.Len(t, missing, len(RequiredBundledFolders))
	})

	t.Run("Bundled All Present", func(t *testing.T) {
		mockClient := new(mocks.Client)
		mockClient.On("BucketExists", mock.Anything, "assets").Return(true, nil)

		for _, folder := range RequiredBundledFolders {
			ch := make(chan minio.ObjectInfo, 1)
			ch <- minio.ObjectInfo{Key: folder + "/"}
			close(ch)
			mockClient.On("ListObjects", mock.Anything, "assets", mock.MatchedBy(func(opts minio.ListObjectsOptions) bool {
				return opts.Prefix == folder+"/"
			})).Return((<-chan minio.ObjectInfo)(ch))
		}

		missing, err := CheckBundled(context.Background(), mockClient, "assets")
		assert.NoError(t, err)
		assert.Len(t, missing, 0)
	})
}

func TestFixBundled(t *testing.T) {
	logger := zap.NewNop()
	mockClient := new(mocks.Client)

	mockClient.On("PutObject", mock.Anything, "assets", mock.Anything, mock.Anything, int64(0), mock.Anything).Return(minio.UploadInfo{}, nil)

	err := FixBundled(context.Background(), mockClient, "assets", logger, []string{"bundled/effect"})
	assert.NoError(t, err)
	mockClient.AssertNumberOfCalls(t, "PutObject", 1)
}
