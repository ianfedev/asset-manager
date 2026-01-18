package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	stdsync "sync"
	"time"

	"asset-manager/core/storage"
	"asset-manager/feature/furniture/models"

	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

// CheckIntegrity performs a high-performance integrity check of bundled furniture.
// Database is mandatory for complete integrity validation across FurniData, Storage, and Database.
func CheckIntegrity(ctx context.Context, client storage.Client, bucket string, db *gorm.DB, emulator string) (*models.Report, error) {
	startTime := time.Now()

	// Validate mandatory database connection
	if db == nil {
		return nil, fmt.Errorf("database connection is required for furniture integrity check")
	}
	if emulator == "" {
		return nil, fmt.Errorf("emulator type is required for furniture integrity check")
	}

	furniData, err := loadFurnitureData(ctx, client, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to load FurnitureData.json: %w", err)
	}

	actualFiles, err := getActualFiles(ctx, client, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to list bundled furniture: %w", err)
	}

	dbAssets, err := getDBAssets(db, emulator)
	if err != nil {
		return nil, fmt.Errorf("failed to load database assets: %w", err)
	}

	report := buildAssetReport(furniData, actualFiles, dbAssets)

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

// buildAssetReport creates a unified asset-centric report by combining data from FurniData, Storage, and Database.
func buildAssetReport(fd *models.FurnitureData, actualFiles map[string]bool, dbAssets map[int]models.DBFurnitureItem) *models.Report {

	// Map to track all unique assets by their base name (without .nitro)
	assetMap := make(map[string]*models.AssetIntegrityItem)

	// Deduplicate FurniData items by ID (Last wins, matching Sync Service behavior)
	activeItems := make(map[int]models.FurnitureItem)

	processDuplicates := func(items []models.FurnitureItem, itemType string) {
		for _, item := range items {
			if existing, exists := activeItems[item.ID]; exists {
				// Log warning (although we don't have logger here, we could add it to report or just overwrite)
				// Since we can't log easily without changing signature, we implicitly accept "Last Wins" strategy.
				// But we should verify if the user needs to know.
				// For now, we strictly follow Sync behavior: Last one overwrites.
				_ = existing
			}
			activeItems[item.ID] = item
		}
	}

	processDuplicates(fd.RoomItemTypes.FurniType, "room")
	processDuplicates(fd.WallItemTypes.FurniType, "wall")

	// Process Active FurniData items
	for _, item := range activeItems {
		// Get base name for file
		baseName := item.ClassName
		if idx := strings.Index(baseName, "*"); idx != -1 {
			baseName = baseName[:idx]
		}
		fileName := baseName + ".nitro"

		// Check if asset already exists in map
		asset, exists := assetMap[fileName]
		if !exists {
			asset = &models.AssetIntegrityItem{
				Name:             fileName,
				FurniDataMissing: false,
				StorageMissing:   true,
				DatabaseMissing:  true,
			}
			assetMap[fileName] = asset
		}

		// Set metadata from FurniData
		asset.ID = item.ID
		asset.ClassName = item.ClassName

		// Validate the FurniData entry
		if msg := item.Validate(); msg != "" {
			asset.Mismatches = append(asset.Mismatches, fmt.Sprintf("FurniData validation: %s", msg))
		}

		// Check database
		if dbItem, found := dbAssets[item.ID]; found {
			asset.DatabaseMissing = false

			// Compare parameters
			// Sanitize FurniData name to match DB storage limitations (Latin1)
			sanitizedName := sanitizeForLatin1(item.Name)
			matches := strings.EqualFold(sanitizedName, dbItem.PublicName)
			if !matches {
				// Check for truncation (common in Arcturus with VARCHAR(56))
				// If DB value is 56 chars and is a prefix of sanitized (case-insensitive), accept it.
				if len(dbItem.PublicName) >= 56 && len(sanitizedName) > len(dbItem.PublicName) {
					if strings.EqualFold(sanitizedName[:len(dbItem.PublicName)], dbItem.PublicName) {
						matches = true
					}
				}
			}

			if !matches {
				asset.Mismatches = append(asset.Mismatches, fmt.Sprintf("name mismatch (FurniData: '%s', DB: '%s')", item.Name, dbItem.PublicName))
			}
			if !strings.EqualFold(item.ClassName, dbItem.ItemName) {
				asset.Mismatches = append(asset.Mismatches, fmt.Sprintf("classname mismatch (FurniData: '%s', DB: '%s')", item.ClassName, dbItem.ItemName))
			}
			if item.XDim != dbItem.Width {
				asset.Mismatches = append(asset.Mismatches, fmt.Sprintf("width mismatch (FurniData: %d, DB: %d)", item.XDim, dbItem.Width))
			}
			if item.YDim != dbItem.Length {
				asset.Mismatches = append(asset.Mismatches, fmt.Sprintf("length mismatch (FurniData: %d, DB: %d)", item.YDim, dbItem.Length))
			}
			if item.CanSitOn != dbItem.CanSit {
				asset.Mismatches = append(asset.Mismatches, fmt.Sprintf("can_sit mismatch (FurniData: %v, DB: %v)", item.CanSitOn, dbItem.CanSit))
			}
			if item.CanStandOn != dbItem.CanWalk {
				asset.Mismatches = append(asset.Mismatches, fmt.Sprintf("can_walk/stand mismatch (FurniData: %v, DB: %v)", item.CanStandOn, dbItem.CanWalk))
			}
			if item.CanLayOn != dbItem.CanLay {
				asset.Mismatches = append(asset.Mismatches, fmt.Sprintf("can_lay mismatch (FurniData: %v, DB: %v)", item.CanLayOn, dbItem.CanLay))
			}
		}

		// Check storage
		if actualFiles[fileName] {
			asset.StorageMissing = false
		}
	}

	// Process storage files not in FurniData (unregistered assets)
	for fileName := range actualFiles {
		if _, exists := assetMap[fileName]; !exists {
			assetMap[fileName] = &models.AssetIntegrityItem{
				Name:             fileName,
				FurniDataMissing: true,
				StorageMissing:   false,
				DatabaseMissing:  true, // We don't know DB status without FurniData ID
			}
		}
	}

	// Process database items not in FurniData (if any)
	for id, dbItem := range dbAssets {
		// Only check if NOT in activeItems
		if _, exists := activeItems[id]; exists {
			continue
		}

		baseName := dbItem.ItemName
		if idx := strings.Index(baseName, "*"); idx != -1 {
			baseName = baseName[:idx]
		}
		fileName := baseName + ".nitro"

		if _, exists := assetMap[fileName]; exists {
			// Already processed
			continue
		}

		// Database-only item (not in FurniData)
		assetMap[fileName] = &models.AssetIntegrityItem{
			ID:               id,
			Name:             fileName,
			ClassName:        dbItem.ItemName,
			FurniDataMissing: true,
			StorageMissing:   !actualFiles[fileName],
			DatabaseMissing:  false,
		}
	}

	// Build report from asset map
	report := &models.Report{
		Assets: make([]models.AssetIntegrityItem, 0, len(assetMap)),
	}

	for _, asset := range assetMap {
		// Calculate mismatches count
		if len(asset.Mismatches) > 0 {
			report.WithMismatches++
		}
		if asset.StorageMissing {
			report.StorageMissing++
		}
		if asset.DatabaseMissing {
			report.DatabaseMissing++
		}
		if asset.FurniDataMissing {
			report.FurniDataMissing++
		}

		// Only include assets that have at least one problem in the details list
		hasIssue := asset.StorageMissing || asset.DatabaseMissing || asset.FurniDataMissing || len(asset.Mismatches) > 0
		if hasIssue {
			report.Assets = append(report.Assets, *asset)
		}
	}

	report.TotalAssets = len(assetMap)

	return report
}

func getActualFiles(ctx context.Context, client storage.Client, bucket string) (map[string]bool, error) {
	actual := make(map[string]bool)
	var mu stdsync.Mutex
	var wg stdsync.WaitGroup

	// Prefixes to scan: a-z, A-Z, 0-9, _, and -
	prefixes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"

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

func getDBAssets(db *gorm.DB, emulator string) (map[int]models.DBFurnitureItem, error) {
	assets := make(map[int]models.DBFurnitureItem)

	switch strings.ToLower(emulator) {
	case "arcturus":
		var items []models.ArcturusItemsBase
		if err := db.Find(&items).Error; err != nil {
			return nil, err
		}
		for _, item := range items {
			assets[item.SpriteID] = item.ToNormalized()
		}
	case "comet":
		var items []models.CometFurniture
		if err := db.Find(&items).Error; err != nil {
			return nil, err
		}
		for _, item := range items {
			assets[item.SpriteID] = item.ToNormalized()
		}
	case "plus":
		var items []models.PlusFurniture
		if err := db.Find(&items).Error; err != nil {
			return nil, err
		}
		for _, item := range items {
			assets[item.SpriteID] = item.ToNormalized()
		}
	default:
		return nil, fmt.Errorf("unsupported emulator: %s", emulator)
	}

	return assets, nil
}

// CheckFurnitureItem performs a detailed integrity check for a single item.
func CheckFurnitureItem(ctx context.Context, client storage.Client, bucket string, db *gorm.DB, emulator string, identifier string) (*models.FurnitureDetailReport, error) {
	report := &models.FurnitureDetailReport{
		IntegrityStatus: "PASS",
	}

	// Clean identifier for DB/Storage search (remove .nitro suffix if present)
	searchIdentifier := identifier
	if strings.HasSuffix(identifier, ".nitro") {
		searchIdentifier = strings.TrimSuffix(identifier, ".nitro")
	}

	// Load FurniData
	furniData, err := loadFurnitureData(ctx, client, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to load FurnitureData: %w", err)
	}

	// Fetch DB Item (if db is present)
	var dbItem *models.DBFurnitureItem
	if db != nil && emulator != "" {
		dbItem, err = GetDBFurnitureItem(db, emulator, searchIdentifier)
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("db lookup failed: %w", err)
		}
	}

	// Find in FurniData
	var item *models.FurnitureItem
	var id int
	isNumericId := false
	if _, err := fmt.Sscanf(identifier, "%d", &id); err == nil && id > 0 {
		isNumericId = true
		report.ID = id
	}

	findInList := func(list []models.FurnitureItem) *models.FurnitureItem {
		for _, idx := range list {
			if isNumericId && idx.ID == id {
				return &idx
			}
			if strings.EqualFold(idx.ClassName, searchIdentifier) {
				return &idx
			}
			if strings.EqualFold(idx.Name, searchIdentifier) {
				return &idx
			}
		}
		return nil
	}

	if found := findInList(furniData.RoomItemTypes.FurniType); found != nil {
		item = found
	} else if found := findInList(furniData.WallItemTypes.FurniType); found != nil {
		item = found
	}

	if item != nil {
		report.InFurniData = true
		report.ID = item.ID
		report.ClassName = item.ClassName
		report.Name = item.Name

		cleanName := item.ClassName
		if idx := strings.Index(cleanName, "*"); idx != -1 {
			cleanName = cleanName[:idx]
		}
		report.NitroFile = cleanName + ".nitro"
	} else {
		report.InFurniData = false
		if isNumericId {
			report.ID = id
		} else {
			report.ClassName = searchIdentifier
			report.NitroFile = searchIdentifier + ".nitro"
		}
	}

	// Retry DB lookup if failed but we have a valid ID from FurniData
	if db != nil && emulator != "" && dbItem == nil && item != nil {
		idStr := fmt.Sprintf("%d", item.ID)
		retryItem, err := GetDBFurnitureItem(db, emulator, idStr)
		if err == nil && retryItem != nil {
			dbItem = retryItem
		}
	}

	// Process DB Result
	if db != nil && emulator != "" {
		if dbItem != nil {
			report.InDB = true
			if report.ID == 0 {
				report.ID = dbItem.ID
			}
			if report.Name == "" {
				report.Name = dbItem.PublicName
			}
			if report.ClassName == "" {
				report.ClassName = dbItem.ItemName
				cleanName := dbItem.ItemName
				if idx := strings.Index(cleanName, "*"); idx != -1 {
					cleanName = cleanName[:idx]
				}
				report.NitroFile = cleanName + ".nitro"
			}

			// Compare if we have both
			if item != nil {
				if item.Name != dbItem.PublicName {
					report.Mismatches = append(report.Mismatches, fmt.Sprintf("Name mismatch: FurniData='%s', DB='%s'", item.Name, dbItem.PublicName))
				}
				if item.ClassName != dbItem.ItemName {
					report.Mismatches = append(report.Mismatches, fmt.Sprintf("ClassName mismatch: FurniData='%s', DB='%s'", item.ClassName, dbItem.ItemName))
				}
				if item.XDim != dbItem.Width {
					report.Mismatches = append(report.Mismatches, fmt.Sprintf("Width mismatch: FurniData=%d, DB=%d", item.XDim, dbItem.Width))
				}
				if item.YDim != dbItem.Length {
					report.Mismatches = append(report.Mismatches, fmt.Sprintf("Length mismatch: FurniData=%d, DB=%d", item.YDim, dbItem.Length))
				}
			}
		} else {
			report.InDB = false
		}
	}

	// Check Storage
	if report.NitroFile != "" {
		filename := report.NitroFile
		pathsToCheck := []string{
			fmt.Sprintf("bundled/furniture/%s", filename),
		}

		if len(filename) > 0 {
			firstChar := string(filename[0])
			pathsToCheck = append(pathsToCheck,
				fmt.Sprintf("bundled/furniture/%s/%s", strings.ToLower(firstChar), filename),
				fmt.Sprintf("bundled/furniture/%s/%s", strings.ToUpper(firstChar), filename),
			)
		}

		foundFile := false
		for _, path := range pathsToCheck {
			opts := minio.ListObjectsOptions{
				Prefix:    path,
				Recursive: false,
				MaxKeys:   1,
			}
			for obj := range client.ListObjects(ctx, bucket, opts) {
				if obj.Err == nil && obj.Key == path {
					foundFile = true
					break
				}
			}
			if foundFile {
				break
			}
		}
		report.FileExists = foundFile
	}

	// Calculate Status
	if !report.InFurniData {
		report.Mismatches = append(report.Mismatches, "Missing in FurniData")
		report.IntegrityStatus = "FAIL"
	}
	if !report.InDB && db != nil {
		report.Mismatches = append(report.Mismatches, "Missing in Database")
		report.IntegrityStatus = "FAIL"
	}
	if !report.FileExists {
		report.Mismatches = append(report.Mismatches, "Missing .nitro file in storage")
		report.IntegrityStatus = "FAIL"
	}
	if len(report.Mismatches) > 0 && report.IntegrityStatus == "PASS" {
		report.IntegrityStatus = "WARNING"
	}

	return report, nil
}
