package batch

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type TestItem struct {
	ID    int    `gamedata:"id"`
	Name  string `gamedata:"name"`
	Value int    `gamedata:"value"`
	Extra string
}

func TestUpdateColumn(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	assert.NoError(t, err)

	logger := zap.NewNop()
	updater := NewBatchUpdater(gormDB, "test_table", "id", logger)

	items := []TestItem{
		{ID: 1, Name: "Item 1", Value: 10},
		{ID: 2, Name: "Item 2", Value: 20},
	}

	t.Run("Update Name Column", func(t *testing.T) {
		mapping := Mapping{StructTag: "name", DBColumn: "public_name"}

		mock.ExpectExec(regexp.QuoteMeta("UPDATE test_table SET public_name = CASE id WHEN ? THEN ? WHEN ? THEN ? END WHERE id IN (?, ?)")).
			WillReturnResult(sqlmock.NewResult(0, 2))

		count, err := updater.UpdateColumn(context.Background(), items, mapping, "id", nil)
		assert.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("Update With Modifier", func(t *testing.T) {
		mapping := Mapping{StructTag: "value", DBColumn: "score"}

		modifier := func(item interface{}, tag string, val interface{}) (interface{}, error) {
			if v, ok := val.(int); ok {
				return v * 2, nil
			}
			return val, nil
		}

		mock.ExpectExec(regexp.QuoteMeta("UPDATE test_table SET score = CASE id WHEN ? THEN ? WHEN ? THEN ? END WHERE id IN (?, ?)")).
			WillReturnResult(sqlmock.NewResult(0, 2))

		count, err := updater.UpdateColumn(context.Background(), items, mapping, "id", modifier)
		assert.NoError(t, err)
		assert.Equal(t, 2, count)
	})
}
