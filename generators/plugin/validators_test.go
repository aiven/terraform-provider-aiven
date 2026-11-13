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

func TestGenValidatorsStringPattern(t *testing.T) {
	t.Parallel()

	const pattern = `^[a-z]+$`

	tests := []struct {
		name string
		item *Item
		want string
	}{
		{
			name: "sensitive string",
			item: &Item{Type: SchemaTypeString, Pattern: pattern, Sensitive: true},
			want: `validators.SensitiveRegexMatches(regexp.MustCompile("^[a-z]+$")`,
		},
		{
			name: "non-sensitive string",
			item: &Item{Type: SchemaTypeString, Pattern: pattern},
			want: `stringvalidator.RegexMatches(regexp.MustCompile("^[a-z]+$")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			codes, err := genValidators(&Definition{}, resourceType, tt.item)
			require.NoError(t, err)
			got := renderCode(t, jen.Var().Id("_").Op("=").Index().Any().Values(codes...))
			assert.Contains(t, got, tt.want)
		})
	}
}
