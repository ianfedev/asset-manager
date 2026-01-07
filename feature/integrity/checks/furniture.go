package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"asset-manager/core/storage"
	"asset-manager/feature/integrity/models"

	"github.com/minio/minio-go/v7"
)

// CheckFurniture performs a high-performance integrity check of bundled furniture.
func CheckFurniture(ctx context.Context, client storage.Client, bucket string) (*models.FurnitureReport, error) {
	startTime := time.Now()

	furniData, err := loadFurnitureData(ctx, client, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to load FurnitureData.json: %w", err)
	}

	expectedFiles := getExpectedFiles(furniData)

	actualFiles, err := getActualFiles(ctx, client, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to list bundled furniture: %w", err)
	}

	report := compareFurniture(expectedFiles, actualFiles)
	report.GeneratedAt = time.Now().Format(time.RFC3339)
	report.ExecutionTime = time.Since(startTime).String()

	return report, nil
}

func loadFurnitureData(ctx context.Context, client storage.Client, bucket string) (*models.FurnitureData, error) {
	objName := "gamedata/FurnitureData.json"

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("bucket %s not found", bucket)
	}

	reader, err := client.GetObject(ctx, bucket, objName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read FurnitureData.json: %w", err)
	}

	var fd models.FurnitureData
	if err := json.Unmarshal(data, &fd); err != nil {
		return nil, fmt.Errorf("failed to parse FurnitureData.json: %w", err)
	}

	return &fd, nil
}

func getExpectedFiles(fd *models.FurnitureData) map[string]bool {
	expected := make(map[string]bool)

	processItems := func(items []models.FurnitureItem) {
		for _, item := range items {
			name := item.ClassName
			if idx := strings.Index(name, "*"); idx != -1 {
				name = name[:idx]
			}
			expected[name+".nitro"] = true
		}
	}

	processItems(fd.RoomItemTypes.FurniType)
	processItems(fd.WallItemTypes.FurniType)

	return expected
}

func getActualFiles(ctx context.Context, client storage.Client, bucket string) (map[string]bool, error) {
	actual := make(map[string]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Prefixes to scan: a-z, A-Z, 0-9, _, and -
	prefixes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"

	// Use a semaphore to limit concurrency if needed, but 37 goroutines is fine.
	// Actually, we can use a channel to collect results to avoid mutex contention if we want,
	// but mutex is simpler and fine for IO bound.

	errCh := make(chan error, 1)

	for _, char := range prefixes {
		prefix := fmt.Sprintf("bundled/furniture/%c", char)
		wg.Add(1)
		go func(p string) {
			defer wg.Done()

			// Check if context canceled or error occurred
			select {
			case <-ctx.Done():
				return
			case <-errCh:
				return
			default:
			}

			opts := minio.ListObjectsOptions{
				Prefix:    p,
				Recursive: true,
			}

			for obj := range client.ListObjects(ctx, bucket, opts) {
				if obj.Err != nil {
					select {
					case errCh <- obj.Err:
					default:
					}
					return
				}

				filename := strings.TrimPrefix(obj.Key, "bundled/furniture/")
				// Simple validation to ensure we caught valid files
				if filename == "" || strings.HasSuffix(filename, "/") {
					continue
				}

				mu.Lock()
				actual[filename] = true
				mu.Unlock()
			}
		}(prefix)
	}

	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
	}

	return actual, nil
}

func compareFurniture(expected map[string]bool, actual map[string]bool) *models.FurnitureReport {
	report := &models.FurnitureReport{
		TotalExpected:      len(expected),
		TotalFound:         len(actual),
		MissingAssets:      make([]string, 0),
		UnregisteredAssets: make([]string, 0),
	}

	// Check missing (In Furnidata but not in Storage)
	for file := range expected {
		if !actual[file] {
			report.MissingAssets = append(report.MissingAssets, file)
		}
	}

	// Check extra (In Storage but not in Furnidata)
	for file := range actual {
		if !expected[file] {
			report.UnregisteredAssets = append(report.UnregisteredAssets, file)
		}
	}

	return report
}
