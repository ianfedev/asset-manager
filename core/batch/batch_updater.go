package batch

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Mapping defines the relationship between a struct field (via tag) and a DB column.
type Mapping struct {
	StructTag string
	DBColumn  string
}

// ValueModifier allows transforming a value before update (e.g. JSON marshaling).
// Return nil to skip update for this item.
type ValueModifier func(item any, tag string, value any) (any, error)

// BatchUpdater handles generic batch updates using CASE statements.
type BatchUpdater struct {
	db        *gorm.DB
	tableName string
	pkColumn  string
	logger    *zap.Logger
	batchSize int
}

// NewBatchUpdater creates a new batch updater instance.
func NewBatchUpdater(db *gorm.DB, tableName, pkColumn string, logger *zap.Logger) *BatchUpdater {
	return &BatchUpdater{
		db:        db,
		tableName: tableName,
		pkColumn:  pkColumn,
		logger:    logger,
		batchSize: 500,
	}
}

// UpdateColumn updates a single column for a slice of items based on the provided mapping.
// It uses the "gamedata" struct tag to locate keys.
func (bu *BatchUpdater) UpdateColumn(ctx context.Context, items any, mapping Mapping, pkTag string, modifier ValueModifier) (int, error) {
	sliceVal := reflect.ValueOf(items)
	if sliceVal.Kind() != reflect.Slice {
		return 0, fmt.Errorf("items must be a slice")
	}

	if sliceVal.Len() == 0 {
		return 0, nil
	}

	// 1. Analyze Struct Type
	elemType := sliceVal.Type().Elem()
	if elemType.Kind() != reflect.Struct {
		return 0, fmt.Errorf("items element must be a struct")
	}

	// Find field indices
	valFieldIdx := -1
	pkFieldIdx := -1

	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		tag := field.Tag.Get("gamedata")

		if tag == mapping.StructTag {
			valFieldIdx = i
		}
		if tag == pkTag {
			pkFieldIdx = i
		}
	}

	if pkFieldIdx == -1 {
		return 0, fmt.Errorf("PK field with tag '%s' not found", pkTag)
	}

	var allIDs []any
	var allValues []any

	for i := 0; i < sliceVal.Len(); i++ {
		itemVal := sliceVal.Index(i)

		// Get PK
		pkVal := itemVal.Field(pkFieldIdx).Interface()

		// Get Value
		var rawValue any
		if valFieldIdx != -1 {
			rawValue = itemVal.Field(valFieldIdx).Interface()
		}

		// Apply modifier
		finalValue := rawValue
		if modifier != nil {
			var err error
			finalValue, err = modifier(itemVal.Interface(), mapping.StructTag, rawValue)
			if err != nil {
				bu.logger.Warn("Modifier error", zap.Error(err))
				continue
			}
		}

		if finalValue == nil {
			continue
		}

		allIDs = append(allIDs, pkVal)
		allValues = append(allValues, finalValue)
	}

	if len(allIDs) == 0 {
		return 0, nil
	}

	// Execute Batches
	totalUpdated := 0
	for i := 0; i < len(allIDs); i += bu.batchSize {
		end := i + bu.batchSize
		if end > len(allIDs) {
			end = len(allIDs)
		}

		chunkIDs := allIDs[i:end]
		chunkValues := allValues[i:end]

		affected, err := bu.executeBatch(chunkIDs, chunkValues, mapping.DBColumn)
		if err != nil {
			return totalUpdated, err
		}
		totalUpdated += int(affected)
	}

	return totalUpdated, nil
}

func (bu *BatchUpdater) executeBatch(ids []any, values []any, column string) (int64, error) {
	var caseWhen []string
	var params []any

	for i := 0; i < len(ids); i++ {
		caseWhen = append(caseWhen, "WHEN ? THEN ?")
		params = append(params, ids[i], values[i])
	}

	for _, id := range ids {
		params = append(params, id)
	}

	placeholders := strings.Repeat("?, ", len(ids)-1) + "?"
	query := fmt.Sprintf("UPDATE %s SET %s = CASE %s %s END WHERE %s IN (%s)",
		bu.tableName, column, bu.pkColumn, strings.Join(caseWhen, " "), bu.pkColumn, placeholders)

	result := bu.db.Exec(query, params...)
	return result.RowsAffected, result.Error
}
