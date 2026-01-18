package integrity

import (
	"context"
	"fmt"

	"asset-manager/core/storage"
	"asset-manager/feature/integrity/checks"

	"asset-manager/feature/furniture/models"
	furnituresync "asset-manager/feature/furniture/sync"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Service handles integrity checks.
type Service struct {
	client   storage.Client
	bucket   string
	logger   *zap.Logger
	db       *gorm.DB
	emulator string
}

// NewService creates a new integrity service.
func NewService(client storage.Client, bucket string, logger *zap.Logger, db *gorm.DB, emulator string) *Service {
	return &Service{
		client:   client,
		bucket:   bucket,
		logger:   logger,
		db:       db,
		emulator: emulator,
	}
}

// Getter methods for sync handler
func (s *Service) GetStorage() storage.Client {
	return s.client
}

func (s *Service) GetBucket() string {
	return s.bucket
}

func (s *Service) GetDB() *gorm.DB {
	return s.db
}

func (s *Service) GetEmulator() string {
	return s.emulator
}

// CheckStructure returns a list of missing folders.
func (s *Service) CheckStructure(ctx context.Context) ([]string, error) {
	return checks.CheckStructure(ctx, s.client, s.bucket)
}

// FixStructure creates the missing folders.
func (s *Service) FixStructure(ctx context.Context, missing []string) error {
	return checks.FixStructure(ctx, s.client, s.bucket, s.logger, missing)
}

// CheckGameData returns a list of missing files in the gamedata folder.
func (s *Service) CheckGameData(ctx context.Context) ([]string, error) {
	return checks.CheckGameData(ctx, s.client, s.bucket)
}

// CheckBundled returns a list of missing bundled folders.
func (s *Service) CheckBundled(ctx context.Context) ([]string, error) {
	return checks.CheckBundled(ctx, s.client, s.bucket)
}

// FixBundled creates the missing bundled folders.
func (s *Service) FixBundled(ctx context.Context, missing []string) error {
	return checks.FixBundled(ctx, s.client, s.bucket, s.logger, missing)
}

// CheckFurniture performs an integrity check on furniture assets.
// Database connection is mandatory for complete integrity validation.
func (s *Service) CheckFurniture(ctx context.Context) (*models.Report, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection is required for furniture integrity check")
	}
	return furnituresync.CheckIntegrity(ctx, s.client, s.bucket, s.db, s.emulator)
}

// CheckServer performs an integrity check on the emulator database schema.
func (s *Service) CheckServer() (*checks.ServerReport, error) {
	if s.db == nil {
		return nil, nil // Or specific error? "Database not connected"
	}
	return checks.CheckServerIntegrity(s.db, s.emulator)
}
