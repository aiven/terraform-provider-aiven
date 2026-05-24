// Copyright (c) 2026 Aiven, Helsinki, Finland. https://aiven.io/

package adapter

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

// TestToTFValue_EmptyCollectionsAreKnown covers the encode path for empty
// List/Set values. Regression for a bug where
// `tftypes.NewValue(<type>, nil)` was returned, which produces a null value.
// When a resource's schema declares the attribute as Required (e.g.
// `accounts` on aiven_byoc_permissions) and the user supplies an empty
// collection, TF rejects the apply with
// `Provider produced inconsistent result after apply ... was
// cty.SetValEmpty(<type>), but now null`.
func TestToTFValue_EmptyCollectionsAreKnown(t *testing.T) {
	t.Parallel()

	strSch := &Schema{Type: SchemaTypeString}

	t.Run("empty list encodes as empty known value, not null", func(t *testing.T) {
		sch := &Schema{Type: SchemaTypeList, Items: strSch}

		got, err := toTFValue(sch, []any{})
		require.NoError(t, err)
		require.False(t, got.IsNull(), "empty list must not be null")
		require.True(t, got.IsKnown(), "empty list must be known")

		// Round-trip: decoding the encoded value must give back []any{}.
		decoded, err := fromTFValueAny(sch, got)
		require.NoError(t, err)
		require.Equal(t, []any{}, decoded)
	})

	t.Run("empty set encodes as empty known value, not null", func(t *testing.T) {
		sch := &Schema{Type: SchemaTypeSet, Items: strSch}

		got, err := toTFValue(sch, []any{})
		require.NoError(t, err)
		require.False(t, got.IsNull(), "empty set must not be null")
		require.True(t, got.IsKnown(), "empty set must be known")

		decoded, err := fromTFValueAny(sch, got)
		require.NoError(t, err)
		require.Equal(t, []any{}, decoded)
	})

	t.Run("non-empty list preserves elements", func(t *testing.T) {
		sch := &Schema{Type: SchemaTypeList, Items: strSch}

		got, err := toTFValue(sch, []any{"a", "b"})
		require.NoError(t, err)
		require.False(t, got.IsNull())

		decoded, err := fromTFValueAny(sch, got)
		require.NoError(t, err)
		require.Equal(t, []any{"a", "b"}, decoded)
	})

	t.Run("non-empty set preserves elements", func(t *testing.T) {
		sch := &Schema{Type: SchemaTypeSet, Items: strSch}

		got, err := toTFValue(sch, []any{"a", "b"})
		require.NoError(t, err)
		require.False(t, got.IsNull())

		decoded, err := fromTFValueAny(sch, got)
		require.NoError(t, err)
		require.Equal(t, []any{"a", "b"}, decoded)
	})

	t.Run("nil value stays null", func(t *testing.T) {
		// Distinct from an empty collection: when the Go-side value is nil,
		// the attribute is unset and we want a null tftypes value. Empty
		// collections (above) and unset (here) are semantically different.
		sch := &Schema{Type: SchemaTypeList, Items: strSch}

		got, err := toTFValue(sch, nil)
		require.NoError(t, err)
		require.True(t, got.IsNull(), "nil input must encode as null")
	})
}

// TestFromTFValueAny_Map covers the decode path for SchemaTypeMap.
// Regression for a bug where SchemaTypeMap and SchemaTypeObject shared
// one case, causing user-supplied map keys to be looked up in sch.Properties
// (which is only populated for Object schemas) and fail with
// `unknown property "<key>"`.
func TestFromTFValueAny_Map(t *testing.T) {
	t.Parallel()

	t.Run("string values decode by key", func(t *testing.T) {
		sch := &Schema{
			Type:  SchemaTypeMap,
			Items: &Schema{Type: SchemaTypeString},
		}
		val := tftypes.NewValue(
			tftypes.Map{ElementType: tftypes.String},
			map[string]tftypes.Value{
				"foo": tftypes.NewValue(tftypes.String, "bar"),
				"baz": tftypes.NewValue(tftypes.String, "qux"),
			},
		)

		got, err := fromTFValueAny(sch, val)
		require.NoError(t, err)
		require.Equal(t, map[string]any{"foo": "bar", "baz": "qux"}, got)
	})

	t.Run("nested object values decode recursively", func(t *testing.T) {
		// Mirrors the `aws_subnets_bastion` / `plan_list.regions` shape:
		// Map<string, Object<availability_zone: string, cidr: string>>.
		inner := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"availability_zone": tftypes.String,
			"cidr":              tftypes.String,
		}}
		sch := &Schema{
			Type: SchemaTypeMap,
			Items: &Schema{
				Type: SchemaTypeObject,
				Properties: map[string]*Schema{
					"availability_zone": {Type: SchemaTypeString},
					"cidr":              {Type: SchemaTypeString},
				},
			},
		}
		val := tftypes.NewValue(
			tftypes.Map{ElementType: inner},
			map[string]tftypes.Value{
				"us_east_1a": tftypes.NewValue(inner, map[string]tftypes.Value{
					"availability_zone": tftypes.NewValue(tftypes.String, "us-east-1a"),
					"cidr":              tftypes.NewValue(tftypes.String, "10.0.0.0/28"),
				}),
				"us_east_1b": tftypes.NewValue(inner, map[string]tftypes.Value{
					"availability_zone": tftypes.NewValue(tftypes.String, "us-east-1b"),
					"cidr":              tftypes.NewValue(tftypes.String, "10.0.0.16/28"),
				}),
			},
		)

		got, err := fromTFValueAny(sch, val)
		require.NoError(t, err)
		require.Equal(t, map[string]any{
			"us_east_1a": map[string]any{
				"availability_zone": "us-east-1a",
				"cidr":              "10.0.0.0/28",
			},
			"us_east_1b": map[string]any{
				"availability_zone": "us-east-1b",
				"cidr":              "10.0.0.16/28",
			},
		}, got)
	})

	t.Run("nil Items is rejected", func(t *testing.T) {
		sch := &Schema{Type: SchemaTypeMap}
		val := tftypes.NewValue(
			tftypes.Map{ElementType: tftypes.String},
			map[string]tftypes.Value{"foo": tftypes.NewValue(tftypes.String, "bar")},
		)

		_, err := fromTFValueAny(sch, val)
		require.Error(t, err)
		require.Contains(t, err.Error(), "map items is nil")
	})

	t.Run("empty map returns empty map", func(t *testing.T) {
		sch := &Schema{
			Type:  SchemaTypeMap,
			Items: &Schema{Type: SchemaTypeString},
		}
		val := tftypes.NewValue(
			tftypes.Map{ElementType: tftypes.String},
			map[string]tftypes.Value{},
		)

		got, err := fromTFValueAny(sch, val)
		require.NoError(t, err)
		require.Equal(t, map[string]any{}, got)
	})
}

// TestFromTFValueAny_Object ensures the Object case still rejects unknown
// keys after the Map/Object split. This is the behavior that existed before
// the decode bug and must be preserved.
func TestFromTFValueAny_Object(t *testing.T) {
	t.Parallel()

	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"name":  tftypes.String,
		"count": tftypes.Number,
	}}

	t.Run("known keys decode", func(t *testing.T) {
		sch := &Schema{
			Type: SchemaTypeObject,
			Properties: map[string]*Schema{
				"name":  {Type: SchemaTypeString},
				"count": {Type: SchemaTypeInt},
			},
		}
		val := tftypes.NewValue(objType, map[string]tftypes.Value{
			"name":  tftypes.NewValue(tftypes.String, "alpha"),
			"count": tftypes.NewValue(tftypes.Number, 3),
		})

		got, err := fromTFValueAny(sch, val)
		require.NoError(t, err)
		require.Equal(t, map[string]any{"name": "alpha", "count": 3}, got)
	})

	t.Run("unknown key errors", func(t *testing.T) {
		// Sanity: an Object schema with fewer Properties than the value has
		// should still raise `unknown property`. This is the symmetric
		// behavior to the Map case — Object enforces the schema; Map accepts
		// anything. Construct an Object value with an extra attribute via
		// a looser Object type.
		looseType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"name":  tftypes.String,
			"extra": tftypes.String,
		}}
		sch := &Schema{
			Type: SchemaTypeObject,
			Properties: map[string]*Schema{
				"name": {Type: SchemaTypeString},
				// note: "extra" deliberately missing
			},
		}
		val := tftypes.NewValue(looseType, map[string]tftypes.Value{
			"name":  tftypes.NewValue(tftypes.String, "alpha"),
			"extra": tftypes.NewValue(tftypes.String, "surprise"),
		})

		_, err := fromTFValueAny(sch, val)
		require.Error(t, err)
		require.Contains(t, err.Error(), `unknown property "extra"`)
	})
}
