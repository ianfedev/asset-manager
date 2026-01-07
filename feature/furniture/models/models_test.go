package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFurnitureItem_Validate(t *testing.T) {
	tests := []struct {
		name     string
		item     FurnitureItem
		expected string
	}{
		{
			name: "Valid Item",
			item: FurnitureItem{
				ID:        1,
				ClassName: "chair",
				Name:      "Chair",
				Category:  "furniture",
			},
			expected: "",
		},
		{
			name: "Valid Item with Color",
			item: FurnitureItem{
				ID:        2,
				ClassName: "chair*1",
				Name:      "Chair",
				Category:  "furniture",
			},
			expected: "",
		},
		{
			name: "Missing ID",
			item: FurnitureItem{
				ClassName: "chair",
				Name:      "Chair",
				Category:  "furniture",
			},
			expected: "missing id",
		},
		{
			name: "Missing ClassName",
			item: FurnitureItem{
				ID:       1,
				Name:     "Chair",
				Category: "furniture",
			},
			expected: "missing classname",
		},
		{
			name: "Missing Name",
			item: FurnitureItem{
				ID:        1,
				ClassName: "chair",
				Category:  "furniture",
			},
			expected: "missing name",
		},
		{
			name: "Missing Category",
			item: FurnitureItem{
				ID:        1,
				ClassName: "chair",
				Name:      "Chair",
			},
			expected: "missing category",
		},
		{
			name: "Invalid ClassName - Too Many Asterisks",
			item: FurnitureItem{
				ID:        1,
				ClassName: "chair*1*2",
				Name:      "Chair",
				Category:  "furniture",
			},
			expected: "invalid classname format: too many asterisks",
		},
		{
			name: "Invalid ClassName - Empty Base",
			item: FurnitureItem{
				ID:        1,
				ClassName: "*1",
				Name:      "Chair",
				Category:  "furniture",
			},
			expected: "invalid classname format: empty base name",
		},
		{
			name: "Invalid ClassName - Empty Color",
			item: FurnitureItem{
				ID:        1,
				ClassName: "chair*",
				Name:      "Chair",
				Category:  "furniture",
			},
			expected: "invalid classname format: empty color index",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.item.Validate()
			assert.Equal(t, tt.expected, result)
		})
	}
}
