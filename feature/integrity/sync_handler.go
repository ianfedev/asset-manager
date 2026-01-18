package integrity

import (
	"asset-manager/core/logger"
	furnituresync "asset-manager/feature/furniture/sync"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// HandleFurnitureSync handles furniture sync requests
// @Summary Sync Furniture
// @Description Synchronize FurniData, Database, and Storage with FurniData as source of truth
// @Tags sync
// @Param confirm query boolean false "Confirm destructive operations (true to execute)"
// @Success 200 {object} map[string]interface{} "Sync Report or Preview"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /sync/furniture [post]
func (h *Handler) HandleFurnitureSync(c *fiber.Ctx) error {
	l := logger.WithRayID(h.service.logger, c)
	l.Info("Furniture sync request received")

	confirmed := c.Query("confirm") == "true"

	// If not confirmed, show preview
	if !confirmed {
		// Run integrity check to show what would be synced
		report, err := h.service.CheckFurniture(c.Context())
		if err != nil {
			l.Error("Integrity check failed", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate sync preview",
			})
		}

		return c.JSON(fiber.Map{
			"preview":           true,
			"message":           "Add ?confirm=true to execute sync",
			"total_assets":      report.TotalAssets,
			"storage_missing":   report.StorageMissing,
			"database_missing":  report.DatabaseMissing,
			"furnidata_missing": report.FurniDataMissing,
			"with_mismatches":   report.WithMismatches,
			"warning":           "This operation will DELETE assets and UPDATE database values",
		})
	}

	// Execute sync
	l.Info("Executing confirmed sync operation")

	// Check if data sync should be skipped (for performance)
	skipDataSync := c.Query("skip-data") == "true"

	syncSvc := furnituresync.NewSyncService(h.service.GetStorage(), h.service.GetBucket(), h.service.GetDB(), h.service.GetEmulator(), l)
	syncReport, err := syncSvc.FullSync(c.Context(), true, skipDataSync)
	if err != nil {
		l.Error("Sync failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	l.Info("Sync completed",
		zap.Int("rows_updated", syncReport.RowsUpdated),
		zap.Int("database_deleted", syncReport.DatabaseDeleted))

	return c.JSON(syncReport)
}
