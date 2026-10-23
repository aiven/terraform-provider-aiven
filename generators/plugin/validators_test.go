package main

import (
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenValidatorsArrayItemEnum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		item *Item
		want string
	}{
		{
			name: "list of strings",
			item: &Item{
				Type:     SchemaTypeArrayOrdered,
				MaxItems: 5,
				Items:    &Item{Type: SchemaTypeString, Enum: []any{"service_name", "HOSTNAME"}},
			},
			want: `listvalidator.ValueStringsAre(stringvalidator.OneOf("HOSTNAME", "service_name"))`,
		},
		{
			name: "set of integers",
			item: &Item{
				Type:  SchemaTypeArray,
				Items: &Item{Type: SchemaTypeInteger, Enum: []any{2, 1}},
			},
			want: `setvalidator.ValueInt64sAre(int64validator.OneOf(1, 2))`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			codes, err := genValidators(&Definition{}, resourceType, tt.item)
			require.NoError(t, err)
			assert.Contains(t, renderCode(t, jen.Var().Id("_").Op("=").Index().Any().Values(codes...)), tt.want)
		})
	}
}

func TestGenValidatorsArrayWithoutItemEnum(t *testing.T) {
	t.Parallel()

	item := &Item{Type: SchemaTypeArrayOrdered, Items: &Item{Type: SchemaTypeString}}
	codes, err := genValidators(&Definition{}, resourceType, item)
	require.NoError(t, err)
	assert.Empty(t, codes)
}
