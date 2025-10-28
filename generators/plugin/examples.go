package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// exampleRoot generates example usage for the TF resource or data source.
func exampleRoot(isResource bool, item *Item) (string, error) {
	t := "data"
	if isResource {
		t = "resource"
	}

	f := hclwrite.NewEmptyFile()
	rootBody := f.Body().AppendNewBlock(t, []string{typeNamePrefix + item.Name, "example"}).Body()
	err := exampleObjectItem(isResource, item, rootBody)
	if err != nil {
		return "", err
	}

	return string(f.Bytes()), nil
}

func sortedKeysPriority(isResource bool, m map[string]*Item) []string {
	keys := slices.Collect(maps.Keys(m))

	slices.SortFunc(keys, func(i, j string) int {
		itemI := m[i]
		itemJ := m[j]

		priorityI := getExampleItemPriority(isResource, itemI)
		priorityJ := getExampleItemPriority(isResource, itemJ)

		if priorityI < priorityJ {
			return -1
		} else if priorityI != priorityJ {
			return 1
		}

		// If both items are ID attributes, sort by position
		if itemI.IDAttributePosition < itemJ.IDAttributePosition {
			return -1
		} else if itemI.IDAttributePosition != itemJ.IDAttributePosition {
			return 1
		}

		// Same priority, sort alphabetically by name
		return strings.Compare(i, j)
	})

	return keys
}

func getExampleItemPriority(isResource bool, item *Item) int {
	switch {
	case item.IsReadOnly(isResource):
		return 2
	case item.InIDAttribute, item.ForceNew:
		return 0
	}
	return 1
}

func exampleObjectItem(isResource bool, item *Item, body *hclwrite.Body) error {
	var seenComputed bool

	for _, k := range sortedKeysPriority(isResource, item.Properties) {
		v := item.Properties[k]

		// Renders COMPUTED FIELDS title before the first computed field
		if v.IsReadOnly(isResource) && item.IsRoot() && !seenComputed {
			seenComputed = true
			comment := hclwrite.Tokens{
				&hclwrite.Token{
					Type:         hclsyntax.TokenComment,
					Bytes:        []byte("// COMPUTED FIELDS"),
					SpacesBefore: 0,
				},
			}
			body.AppendNewline()
			body.AppendUnstructuredTokens(comment)
			body.AppendNewline()
		}

		if v.IsNested() {
			if v.IsArray() {
				v = v.Items
			}

			valBlock := body.AppendNewBlock(k, nil)
			err := exampleObjectItem(isResource, v, valBlock.Body())
			if err != nil {
				return fmt.Errorf("example items error: %w", err)
			}

			continue
		}

		var val cty.Value
		switch {
		case v.IsScalar():
			value, err := exampleScalarItem(v)
			if err != nil {
				return err
			}
			val = value
		case v.IsArray():
			// An array with scalar elements
			value, err := exampleScalarItem(v.Items)
			if err != nil {
				return err
			}

			if v.IsSet() {
				val = cty.SetVal([]cty.Value{value})
			} else {
				val = cty.ListVal([]cty.Value{value})
			}
		case v.IsMapNested():
			// There is no Map Block thing, only Map Attribute.
			// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes/map-nested
			// Currently we support scalars only in map's objects,
			// because otherwise we need to learn to generate nested attributes.
			attrs := make(map[string]cty.Value)
			for kk, vv := range v.Items.Properties {
				if !vv.IsScalar() {
					return fmt.Errorf("unsupported type %s for map %s", vv.Type, v.Path())
				}
				value, err := exampleScalarItem(vv)
				if err != nil {
					return err
				}
				attrs[kk] = value
			}
			val = cty.ObjectVal(map[string]cty.Value{
				"foo": cty.ObjectVal(attrs),
			})
		case v.IsMap():
			value, err := exampleScalarItem(v.Items)
			val = cty.ObjectVal(map[string]cty.Value{
				"foo": value,
			})

			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown property type %q for %s", v.Type, v.Path())
		}

		tokens := hclwrite.TokensForValue(val)
		if isResource && v.ForceNew {
			tokens = append(tokens, &hclwrite.Token{
				Type:  hclsyntax.TokenComment,
				Bytes: []byte("// Force new"),
			})
		}
		body.SetAttributeRaw(k, tokens)
	}

	return nil
}

func exampleScalarItem(item *Item) (cty.Value, error) {
	var anyValue any
	if item.Default != nil {
		anyValue = item.Default
	} else if item.IsEnum() {
		anyValue = item.Enum[0]
	}

	switch item.Type {
	case SchemaTypeString:
		if anyValue == nil {
			// Some placeholders that can improve the example
			placeholders := map[string]any{
				"cidr":            "10.0.0.0/24",
				"email":           "test@example.com",
				"password":        "password",
				"role":            "admin",
				"name":            "test",
				"description":     "test",
				"organization_id": "org1a23f456789",
			}
			anyValue = "foo" // default
			for pattern, example := range placeholders {
				if strings.Contains(item.Name, pattern) {
					anyValue = example
					break
				}
			}
		}
		return cty.StringVal(anyValue.(string)), nil
	case SchemaTypeBoolean:
		if anyValue == nil {
			anyValue = true
		}
		return cty.BoolVal(anyValue.(bool)), nil
	case SchemaTypeInteger:
		if anyValue == nil {
			anyValue = int64(42)
		}
		return cty.NumberIntVal(anyValue.(int64)), nil
	case SchemaTypeNumber:
		if anyValue == nil {
			anyValue = 3.14
		}
		return cty.NumberFloatVal(anyValue.(float64)), nil
	}
	return cty.NilVal, fmt.Errorf("unknown scalar type %s for %s", item.Type, item.Path())
}
