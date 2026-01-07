package checks

import (
	"context"

	"asset-manager/core/storage"

	"go.uber.org/zap"
)

// RequiredBundledFolders lists the bundled asset folders that must exist.
var RequiredBundledFolders = []string{
	"bundled/effect",
	"bundled/figure",
	"bundled/furniture",
	"bundled/generic",
	"bundled/pet",
}

// CheckBundled returns a list of missing bundled folders.
func CheckBundled(ctx context.Context, client storage.Client, bucket string) ([]string, error) {
	return CheckFolders(ctx, client, bucket, RequiredBundledFolders)
}

// FixBundled creates the missing bundled folders.
func FixBundled(ctx context.Context, client storage.Client, bucket string, logger *zap.Logger, missing []string) error {
	return FixFolders(ctx, client, bucket, logger, missing)
}
