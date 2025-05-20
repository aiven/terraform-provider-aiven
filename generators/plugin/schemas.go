package main

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"

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

	// Schema package depends on the entity type.
	pkg := fmtImport(isResource, schemaPackageFmt)
	funcName := "new" + firstUpper(boolEntity(isResource)) + "Schema"
	return jen.Func().
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
	return jen.Qual(fmtImport(isResource, timeoutsPackageFmt), timeoutsMethod).Call(jen.Id("ctx"))
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
		jen.Id("Computed"):            jen.True(),
	}
	if isResource && !idField.Mutable {
		// Make ID field plan modifier to use state for unknown
		// https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#usestateforunknown
		// todo: some resources might change the ID after apply, like aiven_organization_project
		//  in that case, we should not probably use this plan modifier
		attrs[jen.Id("PlanModifiers")] = jen.Index().
			Qual(fmtImport(isResource, planmodifierPackageFmt), "String").
			Values(jen.Qual(getTypedPlanmodifier(SchemaTypeString), "UseStateForUnknown").Call())
	}
	return jen.Qual(fmtImport(isResource, schemaPackageFmt), "StringAttribute").Values(attrs)
}

// genAttributes generates the Attributes field value (map).
func genAttributes(isResource bool, item *Item, extraAttrs jen.Dict) (jen.Dict, error) {
	pkg := fmtImport(isResource, schemaPackageFmt)
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
				value = jen.Qual(typesPackage, v.Element.TFType()+"Type")
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
	pkg := fmtImport(isResource, schemaPackageFmt)
	values, err := genAttributeValues(isResource, item)
	if err != nil {
		return nil, err
	}

	// Renders item or array item
	elem := item.Element
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
				typedPlanmodifier := getTypedPlanmodifier(item.Type)
				values[jen.Id("PlanModifiers")] = jen.Index().Qual(fmtImport(isResource, planmodifierPackageFmt), item.TFType()).
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

	if isResource && !item.IsReadOnly() || !isResource && item.InIDAttribute {
		validators, err := genValidators(item)
		if err != nil {
			return nil, err
		}

		if len(validators) > 0 {
			values[jen.Id("Validators")] = jen.Index().Qual(validatorPackage, item.TFType()).Values(validators...)
		}
	}

	return values, nil
}
