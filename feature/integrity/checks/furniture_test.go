package checks

import (
	"bytes"
	"context"
	"io"
	"testing"

	"asset-manager/core/storage/mocks"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckFurniture(t *testing.T) {
	// Mock JSON data
	mockJSON := `{
		"roomitemtypes": {
			"furnitype": [
				{"id": 1, "classname": "chair"},
				{"id": 2, "classname": "table*1"}
			]
		},
		"wallitemtypes": {
			"furnitype": [
				{"id": 3, "classname": "picture_frame"},
				{"id": 4, "classname": "poster*123"}
			]
		}
	}`

	// Expected files: chair.nitro, table.nitro, picture_frame.nitro, poster.nitro

	t.Run("All Files Present", func(t *testing.T) {
		mockClient := new(mocks.Client)

		mockClient.On("BucketExists", mock.Anything, "assets").Return(true, nil)

		// Mock GetObject
		mockClient.On("GetObject", mock.Anything, "assets", "gamedata/FurnitureData.json", mock.Anything).
			Return(io.NopCloser(bytes.NewReader([]byte(mockJSON))), nil)

		// Mock ListObjects
		// Since implementation is parallel, ListObjects will be called multiple times with prefixes like "bundled/furniture/c" etc.
		// We need to setup the mock to return the correct channel based on the prefix.

		// Helper to create a channel with one item
		createCh := func(key string) <-chan minio.ObjectInfo {
			ch := make(chan minio.ObjectInfo, 1)
			ch <- minio.ObjectInfo{Key: key}
			close(ch)
			return ch
		}

		// Helper for empty channel
		emptyCh := func() <-chan minio.ObjectInfo {
			ch := make(chan minio.ObjectInfo)
			close(ch)
			return ch
		}

		// We use .Maybe() because strict expectation of all 37 calls is tedious and depends on race conditions/scheduler
		// But we must ensure the ones extending to our files ARE called.

		// Matcher for chair (prefix c)
		mockClient.On("ListObjects", mock.Anything, "assets", mock.MatchedBy(func(opts minio.ListObjectsOptions) bool {
			return opts.Prefix == "bundled/furniture/c"
		})).Return(createCh("bundled/furniture/chair.nitro"))

		// Matcher for table (prefix t)
		mockClient.On("ListObjects", mock.Anything, "assets", mock.MatchedBy(func(opts minio.ListObjectsOptions) bool {
			return opts.Prefix == "bundled/furniture/t"
		})).Return(createCh("bundled/furniture/table.nitro"))

		// Matcher for picture_frame (prefix p)
		// Matcher for poster (prefix p) - wait, poster also starts with p.
		// So "bundled/furniture/p" should return BOTH picture_frame and poster.
		// My createCh helper only supports 1. I need to fix that logic.

		// Re-defining return for "p" prefix
		chP := make(chan minio.ObjectInfo, 2)
		chP <- minio.ObjectInfo{Key: "bundled/furniture/picture_frame.nitro"}
		chP <- minio.ObjectInfo{Key: "bundled/furniture/poster.nitro"}
		close(chP)

		// We need to clear previous expectation for P if we added it, but here we are defining it fresh.
		// Actually, testify/mock might evaluate in order.
		// Let's use a generic handler? No, I'll just define specific matchers.

		// Correct expectation for 'p'
		mockClient.On("ListObjects", mock.Anything, "assets", mock.MatchedBy(func(opts minio.ListObjectsOptions) bool {
			return opts.Prefix == "bundled/furniture/p"
		})).Return((<-chan minio.ObjectInfo)(chP))

		// For checking completeness, we should return empty channels for all other calls.
		// Using a fallthrough match?
		mockClient.On("ListObjects", mock.Anything, "assets", mock.Anything).Return(emptyCh())

		report, err := CheckFurniture(context.Background(), mockClient, "assets")
		assert.NoError(t, err)
		assert.Equal(t, 4, report.TotalExpected)
		assert.Equal(t, 4, report.TotalFound)
		assert.Empty(t, report.MissingAssets)
		assert.Empty(t, report.UnregisteredAssets)
	})

	t.Run("Missing and Extra Files", func(t *testing.T) {
		mockClient := new(mocks.Client)

		mockClient.On("BucketExists", mock.Anything, "assets").Return(true, nil)

		mockClient.On("GetObject", mock.Anything, "assets", "gamedata/FurnitureData.json", mock.Anything).
			Return(io.NopCloser(bytes.NewReader([]byte(mockJSON))), nil)

		// Actual: chair.nitro (c), table.nitro (t), extra.nitro (e)

		emptyCh := func() <-chan minio.ObjectInfo {
			ch := make(chan minio.ObjectInfo)
			close(ch)
			return ch
		}

		createCh := func(keys ...string) <-chan minio.ObjectInfo {
			ch := make(chan minio.ObjectInfo, len(keys))
			for _, k := range keys {
				ch <- minio.ObjectInfo{Key: k}
			}
			close(ch)
			return ch
		}

		mockClient.On("ListObjects", mock.Anything, "assets", mock.MatchedBy(func(opts minio.ListObjectsOptions) bool {
			return opts.Prefix == "bundled/furniture/c"
		})).Return(createCh("bundled/furniture/chair.nitro"))

		mockClient.On("ListObjects", mock.Anything, "assets", mock.MatchedBy(func(opts minio.ListObjectsOptions) bool {
			return opts.Prefix == "bundled/furniture/t"
		})).Return(createCh("bundled/furniture/table.nitro"))

		mockClient.On("ListObjects", mock.Anything, "assets", mock.MatchedBy(func(opts minio.ListObjectsOptions) bool {
			return opts.Prefix == "bundled/furniture/e"
		})).Return(createCh("bundled/furniture/extra.nitro"))

		// Fallback for others
		mockClient.On("ListObjects", mock.Anything, "assets", mock.Anything).Return(emptyCh())

		report, err := CheckFurniture(context.Background(), mockClient, "assets")
		assert.NoError(t, err)
		assert.Equal(t, 4, report.TotalExpected)
		assert.Equal(t, 3, report.TotalFound)
		assert.Len(t, report.MissingAssets, 2)
		assert.Contains(t, report.MissingAssets, "picture_frame.nitro")
		assert.Contains(t, report.MissingAssets, "poster.nitro")
		assert.Len(t, report.UnregisteredAssets, 1)
		assert.Contains(t, report.UnregisteredAssets, "extra.nitro")
	})
}
