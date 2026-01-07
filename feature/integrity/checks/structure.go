package checks

import (
	"context"

	"asset-manager/core/storage"

	"go.uber.org/zap"
)

// RequiredFolders lists the folders that must exist in the bucket.
var RequiredFolders = []string{
	"bundled", "c_images", "dcr", "gamedata", "images", "logos", "sounds",
}

// CheckStructure returns a list of missing folders.
func CheckStructure(ctx context.Context, client storage.Client, bucket string) ([]string, error) {
	return CheckFolders(ctx, client, bucket, RequiredFolders)
}

// FixStructure creates the missing folders.
func FixStructure(ctx context.Context, client storage.Client, bucket string, logger *zap.Logger, missing []string) error {
	return FixFolders(ctx, client, bucket, logger, missing)
}
