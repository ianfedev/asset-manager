package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTableNames(t *testing.T) {
	tests := []struct {
		name      string
		model     interface{ TableName() string }
		wantTable string
	}{
		{"Arcturus", ArcturusItemsBase{}, "items_base"},
		{"Comet", CometFurniture{}, "furniture"},
		{"Plus", PlusFurniture{}, "furniture"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantTable, tt.model.TableName())
		})
	}
}
