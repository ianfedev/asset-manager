package sync

import (
	"context"
	"encoding/json"

	"asset-manager/core/batch"
	"asset-manager/feature/furniture/models"

	"go.uber.org/zap"
)

const batchSize = 500

// SyncDataBatch updates ONLY rows with mismatches (from integrity report).
// It leverages the core batch updater to perform CASE updates for mismatched fields.
func (so *SyncOperations) SyncDataBatch(ctx context.Context, furniData *models.FurnitureData, integrityReport *models.Report, logger *zap.Logger) (int, error) {
	mappings, err := so.service.GetParameterMappings()
	if err != nil {
		return 0, err
	}

	tableName := so.service.GetTableName()

	// Filter mismatched items
	mismatchItems := so.buildMismatchItems(furniData, integrityReport)

	logger.Info("Starting targeted data sync",
		zap.Int("total_assets", len(integrityReport.Assets)),
		zap.Int("assets_with_mismatches", len(mismatchItems)),
		zap.Int("total_mappings", len(mappings)))

	if len(mismatchItems) == 0 {
		logger.Info("No mismatches found, skipping data sync")
		return 0, nil
	}

	// Initialize Core Batch Updater
	updater := batch.NewBatchUpdater(so.service.db, tableName, "sprite_id", logger)

	totalUpdated := 0
	for _, mapping := range mappings {
		if mapping.FurniDataParam == "id" {
			continue
		}

		coreMapping := batch.Mapping{
			StructTag: mapping.FurniDataParam,
			DBColumn:  mapping.DBColumn,
		}

		// Use helper createValueModifier to handle transformations
		updated, err := updater.UpdateColumn(ctx, mismatchItems, coreMapping, "id", so.createValueModifier(mapping.DBColumn))
		if err != nil {
			logger.Error("Failed to update column", zap.String("column", mapping.DBColumn), zap.Error(err))
			return totalUpdated, err
		}
		totalUpdated += updated
	}

	return totalUpdated, nil
}

// buildMismatchItems constructs a slice of items that have reported mismatches.
func (so *SyncOperations) buildMismatchItems(furniData *models.FurnitureData, integrityReport *models.Report) []models.FurnitureItem {
	// Build item map
	itemMap := make(map[int]models.FurnitureItem)
	for _, item := range furniData.RoomItemTypes.FurniType {
		itemMap[item.ID] = item
	}
	for _, item := range furniData.WallItemTypes.FurniType {
		itemMap[item.ID] = item
	}

	// Extract IDs with mismatches
	var mismatchItems []models.FurnitureItem
	for _, asset := range integrityReport.Assets {
		// S1009: len(nil slice) is 0, so nil check is redundant.
		if len(asset.Mismatches) > 0 {
			if item, exists := itemMap[asset.ID]; exists {
				mismatchItems = append(mismatchItems, item)
			}
		}
	}
	return mismatchItems
}

// createValueModifier returns a modifier function that handles special type conversions and sanitization.
func (so *SyncOperations) createValueModifier(dbColumn string) batch.ValueModifier {
	return func(item any, tag string, val any) (any, error) {
		// 1. Handle Special Types
		if tag == "partcolors" {
			// val is models.PartColors struct because of the tag
			if pc, ok := val.(struct {
				Color []string `json:"color"`
			}); ok {
				// S1009: len(nil slice) is 0.
				if len(pc.Color) > 0 {
					bytes, _ := json.Marshal(pc)
					return string(bytes), nil
				}
				return nil, nil // Skip update if empty
			}

			// Fallback: Check if it's the specific struct from models directly from item if val assertion fails
			if furniItem, ok := item.(models.FurnitureItem); ok {
				if len(furniItem.PartColors.Color) > 0 {
					bytes, _ := json.Marshal(furniItem.PartColors)
					return string(bytes), nil
				}
			}
			return nil, nil
		}

		// 2. Handle Strings (Sanitize & Truncate)
		if str, ok := val.(string); ok {
			str = sanitizeForLatin1(str)
			maxLen := getColumnMaxLength(dbColumn)
			if maxLen > 0 && len(str) > maxLen {
				str = str[:maxLen]
			}
			return str, nil
		}

		return val, nil
	}
}

// getColumnMaxLength returns the maximum length for a given database column.
func getColumnMaxLength(column string) int {
	limits := map[string]int{
		"item_name":    70,
		"public_name":  56,
		"category":     100,
		"furniline":    100,
		"environment":  100,
		"customparams": 25600,
	}
	if limit, ok := limits[column]; ok {
		return limit
	}
	return 0
}

// sanitizeForLatin1 replaces common special characters that might cause issues in Latin1 columns.
func sanitizeForLatin1(s string) string {
	replacements := map[rune]string{
		0x2018: "'", 0x2019: "'", 0x201C: "\"", 0x201D: "\"",
		0x2013: "-", 0x2014: "-", 0x2026: "...",
	}

	result := []rune{}
	for _, r := range s {
		if r < 256 {
			result = append(result, r)
		} else if replacement, ok := replacements[r]; ok {
			for _, rr := range replacement {
				result = append(result, rr)
			}
		}
	}
	return string(result)
}
