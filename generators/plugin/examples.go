package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/aiven/go-client-codegen/handler/billinggroup"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/samber/lo"
	"github.com/zclconf/go-cty/cty"
)

// exampleRoot generates example usage for the TF resource or data source.
func exampleRoot(def *Definition, entity entityType, item *Item) ([]byte, error) {
	t := "data"
	if entity.isResource() {
		t = "resource"
	}

	f := hclwrite.NewEmptyFile()
	rootBody := f.Body().AppendNewBlock(t, []string{def.typeName, "example"}).Body()
	err := exampleObjectItem(def, entity, item, rootBody, false)
	if err != nil {
		return nil, err
	}

	return f.Bytes(), nil
}

func sortedKeysPriority(def *Definition, entity entityType, item *Item) []string {
	props := item.Properties
	if !entity.isResource() {
		props = item.PropertiesWithoutWO()
	}

	keys := lo.Keys(props)
	slices.SortFunc(keys, func(i, j string) int {
		itemI := props[i]
		itemJ := props[j]

		priorityI := getExampleItemPriority(def, entity, itemI)
		priorityJ := getExampleItemPriority(def, entity, itemJ)

		if priorityI < priorityJ {
			return -1
		} else if priorityI != priorityJ {
			return 1
		}

		// Same priority, sort alphabetically by name
		return strings.Compare(i, j)
	})

	return keys
}

func getExampleItemPriority(def *Definition, entity entityType, item *Item) int {
	score := 100
	required := item.IsRequired(def, entity)
	optional := item.IsOptional(def, entity)

	switch {
	case item.IDAttribute && (optional || required):
		return item.IDAttributePosition
	case required:
		score = 10
	case optional:
		score = 20
	}

	// We want users to use write-only fields, so we prefer them over.
	if entity.isResource() {
		if item.WriteOnly {
			score -= 1
		}

		for _, v := range item.AlsoRequires {
			if item.Parent.Properties[v].WriteOnly {
				score -= 1
				break
			}
		}
	}

	// ID fields are usually prioritized.
	for _, v := range []string{"id", "uuid"} {
		if item.Name == v || strings.HasSuffix(item.Name, "_"+v) {
			score -= 1
			break
		}
	}

	return score
}

// exampleObjectItem renders an object item into the body.
// inComputedBlock - true if the current block is a computed block, so we don't add "Computed" title within a computed block.
func exampleObjectItem(def *Definition, entity entityType, item *Item, body *hclwrite.Body, inComputedBlock bool) error {
	var seenComputed, seenOptional bool
	conflictsWith := make(map[string]bool)

	for i, k := range sortedKeysPriority(def, entity, item) {
		v := item.Properties[k]
		if v.Virtual || v.DeprecationMessage != "" {
			// Don't expose internal virtual properties, like "id"
			// or deprecated properties.
			continue
		}

		// The field conflicts with an rendered property, skip it.
		if conflictsWith[v.Name] {
			continue
		}

		// Stores conflicts with other properties so the example doesn't render them.
		for _, c := range v.ConflictsWith {
			conflictsWith[c] = true
			for _, c := range item.Properties[c].AlsoRequires {
				conflictsWith[c] = true
			}
		}

		// Starts new block: optional or computed.
		// Doesn't add title if it's the first item, except for exactlyOneOf.
		// Doesn't add title if the block is in a computed block.
		var comment string
		exactlyOneOf := !entity.isResource() && def.Datasource != nil && def.Datasource.ExactlyOneOf != nil
		if !inComputedBlock && (i > 0 || exactlyOneOf) {
			if !seenOptional && v.IsOptional(def, entity) {
				seenOptional = true
				comment = "// OPTIONAL FIELDS"

				if exactlyOneOf {
					comment = "// REQUIRED EXACTLY ONE"
				}

			} else if !seenComputed && v.IsReadOnly(def, entity) {
				// Renders COMPUTED FIELDS title before the first computed field
				seenComputed = true
				comment = "/* COMPUTED FIELDS"
			}
		}

		if comment != "" {
			comment := hclwrite.Tokens{
				&hclwrite.Token{
					Type:         hclsyntax.TokenComment,
					Bytes:        []byte(comment),
					SpacesBefore: 0,
				},
			}
			if i > 0 {
				body.AppendNewline()
			}
			body.AppendUnstructuredTokens(comment)
			body.AppendNewline()
		}

		if v.IsNested() {
			if v.IsArray() {
				v = v.Items
			}

			valBlock := body.AppendNewBlock(k, nil)
			err := exampleObjectItem(def, entity, v, valBlock.Body(), seenComputed || inComputedBlock)
			if err != nil {
				return fmt.Errorf("example items error: %w", err)
			}

			continue
		}

		var val cty.Value
		switch {
		case v.IsScalar():
			value, err := exampleScalarItem(def, v)
			if err != nil {
				return err
			}
			val = value
		case v.IsArray():
			// An array with scalar elements
			value, err := exampleScalarItem(def, v.Items)
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
				value, err := exampleScalarItem(def, vv)
				if err != nil {
					return err
				}
				attrs[kk] = value
			}
			val = cty.ObjectVal(map[string]cty.Value{
				"foo": cty.ObjectVal(attrs),
			})
		case v.IsMap():
			value, err := exampleScalarItem(def, v.Items)
			if err != nil {
				return err
			}

			val = cty.ObjectVal(map[string]cty.Value{
				"foo": value,
			})
		default:
			return fmt.Errorf("unknown property type %q for %s", v.Type, v.Path())
		}

		tokens := hclwrite.TokensForValue(val)
		if item.IsRoot() && entity.isResource() && v.ForceNew {
			tokens = append(tokens, &hclwrite.Token{
				Type:  hclsyntax.TokenComment,
				Bytes: []byte("// Force new"),
			})
		}
		body.SetAttributeRaw(k, tokens)
	}

	if !inComputedBlock && seenComputed {
		comment := hclwrite.Tokens{
			&hclwrite.Token{
				Type:         hclsyntax.TokenComment,
				Bytes:        []byte("*/"),
				SpacesBefore: 0,
			},
		}
		body.AppendUnstructuredTokens(comment)
		body.AppendNewline()
	}

	return nil
}

func exampleScalarItem(def *Definition, item *Item) (cty.Value, error) {
	var anyValue any
	switch {
	case !isEmpty(item.Default):
		// Default values are preferred for examples,
		// because "default" usually means "best", while "example" is just a random value.
		anyValue = item.Default
	case !isEmpty(item.Example):
		anyValue = item.Example
	case item.IsEnum():
		anyValue = item.Enum[0]
	}

	switch item.Type {
	case SchemaTypeString:
		if anyValue == nil {
			anyValue = "foo" // Default value
			switch item.Name {
			case "service_name":
				for _, k := range billinggroup.ServiceTypeChoices() {
					// "kafka_connect" is better than "kafka", so it should be preferred
					if strings.Contains(def.typeName, k) {
						anyValue = "my-" + k
						break
					}
				}
			default:
				// Some placeholders that can improve the example
				placeholders := map[string]string{
					"cidr":             "10.0.0.0/24",
					"_ip":              "192.168.1.1",
					"email":            "foo@example.com",
					"password":         "password123",
					"role":             "admin",
					"name":             "foo",
					"description":      "example description",
					"organization_id":  "org1a23f456789",
					"time":             "2021-01-01T00:00:00Z",
					"created_at":       "2021-01-01T00:00:00Z",
					"updated_at":       "2021-01-01T00:00:00Z",
					"project":          "my-project",
					"service_name":     "my-service",
					"billing_group_id": "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d",
				}

				for pattern, example := range placeholders {
					if strings.Contains(item.JSONName, pattern) {
						anyValue = example
						break
					}
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

		// ParseInt handles case when value is a float, e.g. "10.0".
		i, err := strconv.ParseInt(fmt.Sprint(anyValue), 10, 64)
		if err != nil {
			return cty.NilVal, fmt.Errorf(`invalid integer "%v" for %q: %w`, anyValue, item.Path(), err)
		}
		return cty.NumberIntVal(i), nil
	case SchemaTypeNumber:
		if anyValue == nil {
			anyValue = 3.14
		}
		return cty.NumberFloatVal(anyValue.(float64)), nil
	}
	return cty.NilVal, fmt.Errorf("unknown scalar type %q for %q", item.Type, item.Path())
}
