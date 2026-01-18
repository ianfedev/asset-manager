package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"asset-manager/core/config"
	"asset-manager/core/database"
	"asset-manager/core/logger"
	"asset-manager/core/storage"
	furnituresync "asset-manager/feature/furniture/sync"
	"asset-manager/feature/integrity"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize furniture data across FurniData, Database, and Storage",
	Long: `Sync operations use FurniData as the source of truth to:
  • Add missing columns to database schema
  • Update database values to match FurniData
  • Remove assets missing from any source

WARNING: Sync performs destructive operations that cannot be undone.`,
}

var syncFurnitureCmd = &cobra.Command{
	Use:   "furniture",
	Short: "Sync furniture assets across FurniData, Database, and Storage",
	Long: `Performs a complete furniture synchronization:
  1. Runs integrity check to identify issues
  2. Shows summary of changes that will be made
  3. Requests confirmation before proceeding
  4. Executes sync operations (schema, data, asset removal)

This is a DESTRUCTIVE operation.`,
	Run: runFurnitureSync,
}

var syncJSONFlag bool
var skipDataSyncFlag bool

func init() {
	RootCmd.AddCommand(syncCmd)
	syncCmd.AddCommand(syncFurnitureCmd)
	syncFurnitureCmd.Flags().BoolVar(&syncJSONFlag, "json", false, "Output results as JSON")
	syncFurnitureCmd.Flags().BoolVar(&skipDataSyncFlag, "skip-data", false, "Skip data sync (only schema + deletions)")
}

func runFurnitureSync(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()

	// Load config
	cfg, err := config.LoadConfig(".")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create logger
	logg, err := logger.New(&cfg.Log)
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}

	// Create storage client
	store, err := storage.NewClient(cfg.Storage)
	if err != nil {
		logg.Fatal("Failed to create storage client", zap.Error(err))
	}

	// Connect to database (required for sync)
	var db *gorm.DB
	if conn, err := database.Connect(cfg.Database); err != nil {
		logg.Fatal("Database connection required for sync", zap.Error(err))
	} else {
		db = conn
		logg = logg.With(zap.String("server", cfg.Server.Emulator))
	}

	// Create services
	svc := integrity.NewService(store, cfg.Storage.Bucket, logg, db, cfg.Server.Emulator)
	syncSvc := furnituresync.NewSyncService(store, cfg.Storage.Bucket, db, cfg.Server.Emulator, logg)

	// 1. Run integrity check
	logg.Info("Running integrity check...")
	integrityReport, err := svc.CheckFurniture(ctx)
	if err != nil {
		logg.Fatal("Integrity check failed", zap.Error(err))
	}

	// 2. Show preview
	if syncJSONFlag {
		data, _ := json.MarshalIndent(integrityReport, "", "  ")
		fmt.Println(string(data))
		fmt.Println("\nAdd confirm=yes argument to execute")
		return
	} else {
		logg.Info("Furniture Integrity Report",
			zap.Int("TotalAssets", integrityReport.TotalAssets),
			zap.Int("StorageMissing", integrityReport.StorageMissing),
			zap.Int("DatabaseMissing", integrityReport.DatabaseMissing),
			zap.Int("FurniDataMissing", integrityReport.FurniDataMissing),
			zap.Int("WithMismatches", integrityReport.WithMismatches))

		fmt.Println("\n⚠️  WARNING: DESTRUCTIVE OPERATION ⚠️")
		fmt.Println("\nThis sync will:")
		fmt.Printf("  • Add new columns to database schema\n")
		fmt.Printf("  • Update %d rows with mismatched values\n", integrityReport.WithMismatches)
		fmt.Printf("  • DELETE %d items (not in FurniData)\n", integrityReport.FurniDataMissing)

		if integrityReport.DatabaseMissing > 0 {
			fmt.Printf("  • Found %d items missing from database (will be ignored)\n", integrityReport.DatabaseMissing)
		}
		if integrityReport.StorageMissing > 0 {
			fmt.Printf("  • Found %d items missing from storage (will be ignored)\n", integrityReport.StorageMissing)
		}

		fmt.Printf("\nTotal affected: %d assets\n",
			integrityReport.FurniDataMissing+integrityReport.WithMismatches)
		fmt.Println("\nThis action CANNOT be undone. Make sure you have backups.")
	}

	// 3. Request confirmation
	fmt.Print("\nDo you want to proceed? Type 'yes' to continue: ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "yes" {
		logg.Info("Sync cancelled by user")
		return
	}

	// 4. Execute sync
	logg.Info("Starting sync operation...", zap.Bool("skip_data_sync", skipDataSyncFlag))
	syncReport, err := syncSvc.FullSync(ctx, true, skipDataSyncFlag)
	if err != nil {
		logg.Fatal("Sync failed", zap.Error(err))
	}

	// 5. Show results
	if syncJSONFlag {
		data, _ := json.MarshalIndent(syncReport, "", "  ")
		fmt.Println(string(data))
	} else {
		logg.Info("Sync completed successfully",
			zap.Int("RowsUpdated", syncReport.RowsUpdated),
			zap.Int("DatabaseDeleted", syncReport.DatabaseDeleted),
			zap.Int("StorageDeleted", syncReport.StorageDeleted),
			zap.Int("SchemaChanges", len(syncReport.SchemaChanges)),
			zap.String("ExecutionTime", syncReport.ExecutionTime))

		if len(syncReport.SchemaChanges) > 0 {
			fmt.Println("\nSchema Changes:")
			for _, change := range syncReport.SchemaChanges {
				fmt.Printf("  • %s\n", change)
			}
		}

		if len(syncReport.Errors) > 0 {
			fmt.Println("\n⚠️  Errors during sync:")
			for _, errMsg := range syncReport.Errors {
				fmt.Printf("  • %s\n", errMsg)
			}
		}
	}
}
