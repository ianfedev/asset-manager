package checks

import (
	"context"
	"testing"

	"asset-manager/core/storage/mocks"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckGameData(t *testing.T) {
	t.Run("GameData All Missing", func(t *testing.T) {
		mockClient := new(mocks.Client)
		mockClient.On("BucketExists", mock.Anything, "assets").Return(true, nil)

		ch := make(chan minio.ObjectInfo)
		close(ch)
		// For any ListObjects call, return empty channel
		mockClient.On("ListObjects", mock.Anything, "assets", mock.Anything).Return((<-chan minio.ObjectInfo)(ch))

		missing, err := CheckGameData(context.Background(), mockClient, "assets")
		assert.NoError(t, err)
		assert.Len(t, missing, len(RequiredGameDataFiles))
	})

	t.Run("GameData All Present", func(t *testing.T) {
		mockClient := new(mocks.Client)
		mockClient.On("BucketExists", mock.Anything, "assets").Return(true, nil)

		for _, filename := range RequiredGameDataFiles {
			ch := make(chan minio.ObjectInfo, 1)
			ch <- minio.ObjectInfo{Key: "gamedata/" + filename}
			close(ch)

			targetPrefix := "gamedata/" + filename
			mockClient.On("ListObjects", mock.Anything, "assets", mock.MatchedBy(func(opts minio.ListObjectsOptions) bool {
				return opts.Prefix == targetPrefix
			})).Return((<-chan minio.ObjectInfo)(ch))
		}

		missing, err := CheckGameData(context.Background(), mockClient, "assets")
		assert.NoError(t, err)
		assert.Len(t, missing, 0)
	})

	t.Run("Bucket Missing", func(t *testing.T) {
		mockClient := new(mocks.Client)
		mockClient.On("BucketExists", mock.Anything, "assets").Return(false, nil)

		_, err := CheckGameData(context.Background(), mockClient, "assets")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("Bucket Check Error", func(t *testing.T) {
		mockClient := new(mocks.Client)
		mockClient.On("BucketExists", mock.Anything, "assets").Return(false, assert.AnError)

		_, err := CheckGameData(context.Background(), mockClient, "assets")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check bucket existence")
	})
}
