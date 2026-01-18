package sync

import (
	"context"
	"fmt"

	"asset-manager/core/storage"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SyncService handles furniture synchronization operations
type SyncService struct {
	client   storage.Client
	bucket   string
	db       *gorm.DB
	emulator string
	logger   *zap.Logger
}

// NewSyncService creates a new sync service
func NewSyncService(client storage.Client, bucket string, db *gorm.DB, emulator string, logger *zap.Logger) *SyncService {
	return &SyncService{
		client:   client,
		bucket:   bucket,
		db:       db,
		emulator: emulator,
		logger:   logger,
	}
}

// SyncReport contains the results of a sync operation
type SyncReport struct {
	SchemaChanges    []string `json:"schema_changes"`
	RowsUpdated      int      `json:"rows_updated"`
	AssetsDeleted    int      `json:"assets_deleted"`
	StorageDeleted   int      `json:"storage_deleted"`
	DatabaseDeleted  int      `json:"database_deleted"`
	FurniDataDeleted int      `json:"furnidata_deleted"`
	ExecutionTime    string   `json:"execution_time"`
	Errors           []string `json:"errors,omitempty"`
}

// ParameterMapping represents a FurniData parameter to DB column mapping
type ParameterMapping struct {
	FurniDataParam string
	DBColumn       string
	DBType         string
	DefaultValue   string
	IsNewColumn    bool
}

// GetParameterMappings returns parameter mappings for the given emulator
func (s *SyncService) GetParameterMappings() ([]ParameterMapping, error) {
	switch s.emulator {
	case "arcturus":
		return getArcturusMappings(), nil
	case "comet":
		return getCometMappings(), nil
	case "plus":
		return getPlusMappings(), nil
	default:
		return nil, fmt.Errorf("unsupported emulator: %s", s.emulator)
	}
}

// getArcturusMappings returns parameter mappings for Arcturus (items_base table)
func getArcturusMappings() []ParameterMapping {
	return []ParameterMapping{
		// Existing mappings
		{FurniDataParam: "id", DBColumn: "sprite_id", DBType: "INT", IsNewColumn: false},
		{FurniDataParam: "classname", DBColumn: "item_name", DBType: "VARCHAR(70)", IsNewColumn: false},
		{FurniDataParam: "name", DBColumn: "public_name", DBType: "VARCHAR(56)", IsNewColumn: false},
		{FurniDataParam: "xdim", DBColumn: "width", DBType: "INT", DefaultValue: "1", IsNewColumn: false},
		{FurniDataParam: "ydim", DBColumn: "length", DBType: "INT", DefaultValue: "1", IsNewColumn: false},
		{FurniDataParam: "cansiton", DBColumn: "allow_sit", DBType: "TINYINT(1)", DefaultValue: "0", IsNewColumn: false},
		{FurniDataParam: "canlayon", DBColumn: "allow_lay", DBType: "TINYINT(1)", DefaultValue: "0", IsNewColumn: false},
		{FurniDataParam: "canstandon", DBColumn: "allow_walk", DBType: "TINYINT(1)", DefaultValue: "0", IsNewColumn: false},
		{FurniDataParam: "customparams", DBColumn: "customparams", DBType: "VARCHAR(25600)", DefaultValue: "NULL", IsNewColumn: false},

		// New columns to be added
		{FurniDataParam: "description", DBColumn: "description", DBType: "TEXT", DefaultValue: "NULL", IsNewColumn: true},
		{FurniDataParam: "revision", DBColumn: "revision", DBType: "INT", DefaultValue: "0", IsNewColumn: true},
		{FurniDataParam: "category", DBColumn: "category", DBType: "VARCHAR(100)", DefaultValue: "''", IsNewColumn: true},
		{FurniDataParam: "offerid", DBColumn: "offerid", DBType: "INT", DefaultValue: "0", IsNewColumn: true},
		{FurniDataParam: "buyout", DBColumn: "buyout", DBType: "TINYINT(1)", DefaultValue: "0", IsNewColumn: true},
		{FurniDataParam: "rentofferid", DBColumn: "rentofferid", DBType: "INT", DefaultValue: "0", IsNewColumn: true},
		{FurniDataParam: "rentbuyout", DBColumn: "rentbuyout", DBType: "TINYINT(1)", DefaultValue: "0", IsNewColumn: true},
		{FurniDataParam: "bc", DBColumn: "bc", DBType: "TINYINT(1)", DefaultValue: "0", IsNewColumn: true},
		{FurniDataParam: "excludeddynamic", DBColumn: "excludeddynamic", DBType: "TINYINT(1)", DefaultValue: "0", IsNewColumn: true},
		{FurniDataParam: "furniline", DBColumn: "furniline", DBType: "VARCHAR(100)", DefaultValue: "''", IsNewColumn: true},
		{FurniDataParam: "environment", DBColumn: "environment", DBType: "VARCHAR(100)", DefaultValue: "''", IsNewColumn: true},
		{FurniDataParam: "adurl", DBColumn: "adurl", DBType: "TEXT", DefaultValue: "NULL", IsNewColumn: true},
		{FurniDataParam: "defaultdir", DBColumn: "defaultdir", DBType: "INT", DefaultValue: "0", IsNewColumn: true},
		{FurniDataParam: "partcolors", DBColumn: "partcolors", DBType: "TEXT", DefaultValue: "NULL", IsNewColumn: true},
		{FurniDataParam: "specialtype", DBColumn: "furni_specialtype", DBType: "INT", DefaultValue: "0", IsNewColumn: true},
	}
}

// getCometMappings returns parameter mappings for Comet (furniture table)
func getCometMappings() []ParameterMapping {
	return []ParameterMapping{
		// Existing mappings
		{FurniDataParam: "id", DBColumn: "sprite_id", DBType: "INT", IsNewColumn: false},
		{FurniDataParam: "classname", DBColumn: "item_name", DBType: "VARCHAR(255)", IsNewColumn: false},
		{FurniDataParam: "name", DBColumn: "public_name", DBType: "VARCHAR(255)", IsNewColumn: false},
		{FurniDataParam: "xdim", DBColumn: "width", DBType: "INT", DefaultValue: "1", IsNewColumn: false},
		{FurniDataParam: "ydim", DBColumn: "length", DBType: "INT", DefaultValue: "1", IsNewColumn: false},
		{FurniDataParam: "cansiton", DBColumn: "can_sit", DBType: "ENUM('0','1')", DefaultValue: "'0'", IsNewColumn: false},
		{FurniDataParam: "canlayon", DBColumn: "can_lay", DBType: "ENUM('0','1')", DefaultValue: "'0'", IsNewColumn: false},
		{FurniDataParam: "canstandon", DBColumn: "is_walkable", DBType: "ENUM('0','1')", DefaultValue: "'0'", IsNewColumn: false},
		{FurniDataParam: "revision", DBColumn: "revision", DBType: "INT", DefaultValue: "45554", IsNewColumn: false},
		{FurniDataParam: "description", DBColumn: "description", DBType: "VARCHAR(255)", IsNewColumn: false},
		{FurniDataParam: "partcolors", DBColumn: "colors", DBType: "LONGTEXT", DefaultValue: "NULL", IsNewColumn: false},

		// New columns to be added
		{FurniDataParam: "category", DBColumn: "category", DBType: "VARCHAR(100)", DefaultValue: "''", IsNewColumn: true},
		{FurniDataParam: "offerid", DBColumn: "offerid", DBType: "INT", DefaultValue: "0", IsNewColumn: true},
		{FurniDataParam: "buyout", DBColumn: "buyout", DBType: "ENUM('0','1')", DefaultValue: "'0'", IsNewColumn: true},
		{FurniDataParam: "rentofferid", DBColumn: "rentofferid", DBType: "INT", DefaultValue: "0", IsNewColumn: true},
		{FurniDataParam: "rentbuyout", DBColumn: "rentbuyout", DBType: "ENUM('0','1')", DefaultValue: "'0'", IsNewColumn: true},
		{FurniDataParam: "bc", DBColumn: "bc", DBType: "ENUM('0','1')", DefaultValue: "'0'", IsNewColumn: true},
		{FurniDataParam: "excludeddynamic", DBColumn: "excludeddynamic", DBType: "ENUM('0','1')", DefaultValue: "'0'", IsNewColumn: true},
		{FurniDataParam: "furniline", DBColumn: "furniline", DBType: "VARCHAR(100)", DefaultValue: "''", IsNewColumn: true},
		{FurniDataParam: "environment", DBColumn: "environment", DBType: "VARCHAR(100)", DefaultValue: "''", IsNewColumn: true},
		{FurniDataParam: "adurl", DBColumn: "adurl", DBType: "TEXT", DefaultValue: "NULL", IsNewColumn: true},
		{FurniDataParam: "defaultdir", DBColumn: "defaultdir", DBType: "INT", DefaultValue: "0", IsNewColumn: true},
		{FurniDataParam: "customparams", DBColumn: "customparams", DBType: "TEXT", DefaultValue: "NULL", IsNewColumn: true},
		{FurniDataParam: "specialtype", DBColumn: "furni_specialtype", DBType: "INT", DefaultValue: "0", IsNewColumn: true},
		{FurniDataParam: "rare", DBColumn: "is_rare", DBType: "ENUM('0','1')", DefaultValue: "'0'", IsNewColumn: true},
	}
}

// getPlusMappings returns parameter mappings for Plus (furniture table)
func getPlusMappings() []ParameterMapping {
	return []ParameterMapping{
		// Existing mappings
		{FurniDataParam: "id", DBColumn: "sprite_id", DBType: "INT", IsNewColumn: false},
		{FurniDataParam: "classname", DBColumn: "item_name", DBType: "VARCHAR(255)", IsNewColumn: false},
		{FurniDataParam: "name", DBColumn: "public_name", DBType: "VARCHAR(255)", IsNewColumn: false},
		{FurniDataParam: "xdim", DBColumn: "width", DBType: "INT", DefaultValue: "1", IsNewColumn: false},
		{FurniDataParam: "ydim", DBColumn: "length", DBType: "INT", DefaultValue: "1", IsNewColumn: false},
		{FurniDataParam: "cansiton", DBColumn: "can_sit", DBType: "TINYINT(1)", DefaultValue: "0", IsNewColumn: false},
		{FurniDataParam: "canstandon", DBColumn: "is_walkable", DBType: "TINYINT(1)", DefaultValue: "0", IsNewColumn: false},
		{FurniDataParam: "rare", DBColumn: "is_rare", DBType: "TINYINT(1)", DefaultValue: "0", IsNewColumn: false},

		// New columns to be added
		{FurniDataParam: "description", DBColumn: "description", DBType: "TEXT", DefaultValue: "NULL", IsNewColumn: true},
		{FurniDataParam: "revision", DBColumn: "revision", DBType: "INT", DefaultValue: "0", IsNewColumn: true},
		{FurniDataParam: "category", DBColumn: "category", DBType: "VARCHAR(100)", DefaultValue: "''", IsNewColumn: true},
		{FurniDataParam: "offerid", DBColumn: "offerid", DBType: "INT", DefaultValue: "0", IsNewColumn: true},
		{FurniDataParam: "buyout", DBColumn: "buyout", DBType: "TINYINT(1)", DefaultValue: "0", IsNewColumn: true},
		{FurniDataParam: "rentofferid", DBColumn: "rentofferid", DBType: "INT", DefaultValue: "0", IsNewColumn: true},
		{FurniDataParam: "rentbuyout", DBColumn: "rentbuyout", DBType: "TINYINT(1)", DefaultValue: "0", IsNewColumn: true},
		{FurniDataParam: "bc", DBColumn: "bc", DBType: "TINYINT(1)", DefaultValue: "0", IsNewColumn: true},
		{FurniDataParam: "excludeddynamic", DBColumn: "excludeddynamic", DBType: "TINYINT(1)", DefaultValue: "0", IsNewColumn: true},
		{FurniDataParam: "furniline", DBColumn: "furniline", DBType: "VARCHAR(100)", DefaultValue: "''", IsNewColumn: true},
		{FurniDataParam: "environment", DBColumn: "environment", DBType: "VARCHAR(100)", DefaultValue: "''", IsNewColumn: true},
		{FurniDataParam: "adurl", DBColumn: "adurl", DBType: "TEXT", DefaultValue: "NULL", IsNewColumn: true},
		{FurniDataParam: "defaultdir", DBColumn: "defaultdir", DBType: "INT", DefaultValue: "0", IsNewColumn: true},
		{FurniDataParam: "partcolors", DBColumn: "partcolors", DBType: "TEXT", DefaultValue: "NULL", IsNewColumn: true},
		{FurniDataParam: "customparams", DBColumn: "customparams", DBType: "TEXT", DefaultValue: "NULL", IsNewColumn: true},
		{FurniDataParam: "specialtype", DBColumn: "furni_specialtype", DBType: "INT", DefaultValue: "0", IsNewColumn: true},
		{FurniDataParam: "canlayon", DBColumn: "can_lay", DBType: "TINYINT(1)", DefaultValue: "0", IsNewColumn: true},
	}
}

// FullSync performs a complete synchronization operation
func (s *SyncService) FullSync(ctx context.Context, confirmed bool, skipDataSync bool) (*SyncReport, error) {
	if !confirmed {
		return nil, fmt.Errorf("sync operation requires confirmation")
	}

	ops := NewSyncOperations(s)
	return ops.PerformFullSync(ctx, skipDataSync)
}

// GetTableName returns the table name for the emulator
func (s *SyncService) GetTableName() string {
	switch s.emulator {
	case "arcturus":
		return "items_base"
	case "comet", "plus":
		return "furniture"
	default:
		return ""
	}
}
