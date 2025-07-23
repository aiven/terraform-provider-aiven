package main

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/samber/lo"
	"github.com/zclconf/go-cty/cty"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

func genSchema(isBeta, isResource bool, item *Item, idField *IDAttribute) (jen.Code, error) {
	extrAttrs := jen.Dict{
		jen.Lit("id"): genIDField(isResource, idField),
	}

	attrs, err := genAttributes(isResource, item, extrAttrs)
	if err != nil {
		return nil, err
	}

	// We use DeprecationMessage instead of DeprecationMessage
	// to render links and use Markdown formatting
	if item.DeprecationMessage != "" {
		attrs[jen.Id("DeprecationMessage")] = jen.Lit(item.DeprecationMessage)
	}

	desc := fmtDescription(isResource, item)
	if isBeta {
		desc = userconfig.Desc(desc).AvailabilityType(userconfig.Beta).Build()
	}
	attrs[jen.Id("MarkdownDescription")] = jen.Lit(desc)

	example, err := exampleRoot(isResource, item)
	if err != nil {
		return nil, fmt.Errorf("example error: %w", err)
	}

	// Schema package depends on the entity type.
	pkg := entityImport(isResource, schemaPackageFmt)
	funcName := "new" + firstUpper(boolEntity(isResource)) + "Schema"
	return jen.
		Comment(funcName+":\n"+example).
		Line().
		Func().
		Id(funcName).
		Params(jen.Id("ctx").Qual("context", "Context")).
		Qual(pkg, "Schema").
		Block(jen.Return(jen.Qual(pkg, "Schema").Values(attrs))), nil
}

// genTimeoutsField generates the timeouts field for the schema.
// The result depends on the entity type.
func genTimeoutsField(isResource bool) jen.Code {
	timeoutsMethod := "Block"
	if isResource {
		timeoutsMethod = "BlockAll"
	}
	return jen.Qual(entityImport(isResource, timeoutsPackageFmt), timeoutsMethod).Call(jen.Id("ctx"))
}

// genIDField generates the ID field for the schema.
// The result depends on the entity type.
func genIDField(isResource bool, idField *IDAttribute) jen.Code {
	description := idField.Description
	if description == "" {
		fields := lo.Map(idField.Compose, func(v string, _ int) string {
			return fmt.Sprintf("`%s`", v)
		})
		switch len(idField.Compose) {
		case 0:
		case 1:
			description = fmt.Sprintf("Resource ID, equal to %s.", fields[0])
		default:
			description = fmt.Sprintf("Resource ID, a composite of %s IDs.", joinCommaAnd(fields))
		}
	}

	attrs := jen.Dict{
		jen.Id("MarkdownDescription"): jen.Lit(description),
	}

	if !isResource && len(idField.Compose) == 1 && idField.Compose[0] == "id" {
		// If datasource id is literally "id", it is required.
		attrs[jen.Id("Required")] = jen.True()
	} else {
		attrs[jen.Id("Computed")] = jen.True()
	}

	if isResource && !idField.Mutable {
		// Make ID field plan modifier to use state for unknown
		// https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#usestateforunknown
		// todo: some resources might change the ID after apply, like aiven_organization_project
		//  in that case, we should not probably use this plan modifier
		attrs[jen.Id("PlanModifiers")] = jen.Index().
			Qual(entityImport(isResource, planmodifierPackageFmt), "String").
			Values(jen.Qual(getTypedImport(SchemaTypeString, planmodifierTypedImport), "UseStateForUnknown").Call())
	}
	return jen.Qual(entityImport(isResource, schemaPackageFmt), "StringAttribute").Values(attrs)
}

// genAttributes generates the Attributes field value (map).
func genAttributes(isResource bool, item *Item, extraAttrs jen.Dict) (jen.Dict, error) {
	pkg := entityImport(isResource, schemaPackageFmt)
	attrs := make(jen.Dict)
	blocks := make(jen.Dict)

	for _, k := range sortedKeys(item.Properties) {
		if k == "id" && item.IsRoot() {
			// The id field is add by genIDField
			continue
		}

		v := item.Properties[k]
		key := jen.Lit(k)
		if v.IsNested() {
			value, err := genSetNestedBlock(isResource, v)
			if err != nil {
				return nil, err
			}
			blocks[key] = value
		} else {
			var value *jen.Statement
			switch v.Type {
			case SchemaTypeArray, SchemaTypeObject:
				value = jen.Qual(typesPackage, v.Items.TFType()+"Type")
			}

			values, err := genAttributeValues(isResource, v)
			if err != nil {
				return nil, err
			}

			if value != nil {
				values[jen.Id("ElementType")] = value
			}

			attrs[key] = jen.Qual(pkg, v.TFType()+"Attribute").Values(values)
		}
	}

	if item.IsRoot() {
		blocks[jen.Lit("timeouts")] = genTimeoutsField(isResource)
	}

	for k, v := range extraAttrs {
		attrs[k] = v
	}

	nested := make(jen.Dict)
	if len(blocks) > 0 {
		nested[jen.Id("Blocks")] = jen.Map(jen.String()).Qual(pkg, "Block").Values(blocks)
	}

	if len(attrs) > 0 {
		nested[jen.Id("Attributes")] = jen.Map(jen.String()).Qual(pkg, "Attribute").Values(attrs)
	}

	return nested, nil
}

func genSetNestedBlock(isResource bool, item *Item) (jen.Code, error) {
	pkg := entityImport(isResource, schemaPackageFmt)
	values, err := genAttributeValues(isResource, item)
	if err != nil {
		return nil, err
	}

	// Renders item or array item
	elem := item.Items
	if elem == nil {
		elem = item
	}

	nested, err := genAttributes(isResource, elem, nil)
	if err != nil {
		return nil, err
	}

	t := "SetNestedBlock"
	if item.IsObject() {
		t = "ListNestedBlock"
	}

	values[jen.Id("NestedObject")] = jen.Qual(pkg, "NestedBlockObject").Values(nested)
	return jen.Qual(pkg, t).Values(values), nil
}

func genAttributeValues(isResource bool, item *Item) (jen.Dict, error) {
	values := jen.Dict{}
	description := fmtDescription(isResource, item)
	if description != "" {
		values[jen.Id("MarkdownDescription")] = jen.Lit(description)
	}

	if item.DeprecationMessage != "" {
		values[jen.Id("DeprecationMessage")] = jen.Lit(item.DeprecationMessage)
	}

	if item.Sensitive {
		values[jen.Id("Sensitive")] = jen.True()
	}

	if !item.IsNested() {
		if isResource {
			if item.ForceNew {
				typedPlanmodifier := getTypedImport(item.Type, planmodifierTypedImport)
				values[jen.Id("PlanModifiers")] = jen.Index().Qual(entityImport(isResource, planmodifierPackageFmt), item.TFType()).
					Values(jen.Qual(typedPlanmodifier, "RequiresReplace").Call())
			}

			if item.Required {
				values[jen.Id("Required")] = jen.True()
			}

			if item.Optional {
				values[jen.Id("Optional")] = jen.True()
			}

			if item.Computed {
				values[jen.Id("Computed")] = jen.True()
			}
		} else {
			if item.InIDAttribute {
				values[jen.Id("Required")] = jen.True()
			} else {
				values[jen.Id("Computed")] = jen.True()
			}
		}
	}

	if !item.IsReadOnly(isResource) {
		validators, err := genValidators(item)
		if err != nil {
			return nil, err
		}

		if len(validators) > 0 {
			values[jen.Id("Validators")] = jen.Index().Qual(validatorPackage, item.TFType()).Values(validators...)
		}

		if item.Default != nil {
			values[jen.Id("Default")] = jen.Qual(getTypedImport(item.Type, defaultsTypedImport), "String").Call(jen.Lit(item.Default))
		}
	}

	return values, nil
}

// exampleRoot generates example usage for the TF resource or data source.
func exampleRoot(isResource bool, item *Item) (string, error) {
	t := "data"
	if isResource {
		t = "resource"
	}

	f := hclwrite.NewEmptyFile()
	rootBody := f.Body().AppendNewBlock(t, []string{"aiven_" + item.Name, "example"}).Body()
	err := exampleObjectItem(isResource, item, rootBody)
	if err != nil {
		return "", err
	}

	return string(f.Bytes()), nil
}

func sortedKeysPriority(isResource bool, m map[string]*Item) []string {
	keys := slices.Collect(maps.Keys(m))
	sort.Slice(keys, func(i, j int) bool {
		itemI := m[keys[i]]
		itemJ := m[keys[j]]

		priorityI := getExampleItemPriority(isResource, itemI)
		priorityJ := getExampleItemPriority(isResource, itemJ)

		if priorityI != priorityJ {
			return priorityI < priorityJ
		}

		// If both items are ID attributes, sort by position
		if itemI.IDAttributePosition != itemJ.IDAttributePosition {
			return itemI.IDAttributePosition < itemJ.IDAttributePosition
		}

		// Same priority, sort alphabetically by name
		return strings.Compare(itemI.Name, itemJ.Name) < 0
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
				return err
			}

			continue
		}

		var val cty.Value
		switch v.Type {
		case SchemaTypeString, SchemaTypeBoolean, SchemaTypeInteger, SchemaTypeNumber:
			value, err := exampleScalarItem(v)
			if err != nil {
				return err
			}
			val = value
		case SchemaTypeArray:
			// An array with scalar elements
			value, err := exampleScalarItem(v.Items)
			if err != nil {
				return err
			}
			val = cty.ListVal([]cty.Value{value})
		case SchemaTypeObject:
			value, err := exampleScalarItem(v.Items)
			if err != nil {
				return err
			}
			val = cty.ObjectVal(map[string]cty.Value{
				"foo": value,
			})
		default:
			return fmt.Errorf("unsupported type %s for %s", v.Type, v.Path())
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
	return cty.NilVal, fmt.Errorf("unsupported type %s for %s", item.Type, item.Path())
}
