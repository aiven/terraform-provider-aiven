package main

import (
	"fmt"
	"slices"

	"github.com/dave/jennifer/jen"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

const schemaSuffix = "Schema"

func genSchema(entity entityType, item *Item, def *Definition) (jen.Code, error) {
	attrs, err := genAttributes(entity.isResource(), item, def)
	if err != nil {
		return nil, err
	}

	// We use DeprecationMessage instead of DeprecationMessage
	// to render links and use Markdown formatting
	if item.DeprecationMessage != "" {
		attrs[jen.Id("DeprecationMessage")] = jen.Lit(item.DeprecationMessage)
	}

	desc := fmtDescription(entity.isResource(), item)
	if def.Beta {
		desc = userconfig.Desc(desc).AvailabilityType(userconfig.Beta).Build()
	}
	attrs[jen.Id("MarkdownDescription")] = jen.Lit(desc)

	if entity.isResource() && def.Version != nil {
		// Only resources have Version field
		attrs[jen.Id("Version")] = jen.Lit(*def.Version)
	}

	example, err := exampleRoot(entity.isResource(), item)
	if err != nil {
		return nil, fmt.Errorf("example error: %w", err)
	}

	// Schema package depends on the entity type.
	pkg := entityImport(entity.isResource(), schemaPackageFmt)
	funcName := string(entity) + schemaSuffix
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
// The result depends on the entity type and whether legacy timeouts are used.
func genTimeoutsField(isResource bool, def *Definition) jen.Code {
	pkg := entityImport(isResource, timeoutsPackageFmt)
	method := "Block"
	if isResource {
		method = "BlockAll"
		if def.LegacyTimeouts {
			pkg = legacyTimeoutsPackage
		}
	}
	return jen.Qual(pkg, method).Call(jen.Id("ctx"))
}

// genAttributes generates the Attributes field value (map).
func genAttributes(isResource bool, item *Item, def *Definition) (jen.Dict, error) {
	pkg := entityImport(isResource, schemaPackageFmt)
	attrs := make(jen.Dict)
	blocks := make(jen.Dict)

	for _, k := range sortedKeys(item.Properties) {
		v := item.Properties[k]
		key := jen.Lit(k)
		switch {
		case v.IsMapNested():
			// There is no Map Block thing, only Map Attribute.
			// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes/map-nested
			value, err := genMapNestedAttribute(isResource, v)
			if err != nil {
				return nil, err
			}
			attrs[key] = value
		case v.IsNested():
			value, err := genSetNestedBlock(isResource, v)
			if err != nil {
				return nil, err
			}
			blocks[key] = value
		default:
			var value *jen.Statement
			if v.Items != nil {
				value = jen.Qual(typesPackage, v.Items.TFType()+"Type")
			}

			values, err := genAttributeValues(isResource, v)
			if err != nil {
				return nil, err
			}

			if value != nil {
				values["ElementType"] = value
			}

			if !isResource && item.IsRoot() && slices.Contains(def.Datasource.ExactlyOneOf, k) {
				delete(values, "Required")
				values["Optional"] = jen.True()
				values["Computed"] = jen.True()
			}

			attrs[key] = jen.Qual(pkg, v.TFType()+"Attribute").Values(dictFromMap(values, false))
		}
	}

	if item.IsRoot() && def != nil {
		blocks[jen.Lit("timeouts")] = genTimeoutsField(isResource, def)
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

func genMapNestedAttribute(isResource bool, item *Item) (jen.Code, error) {
	pkg := entityImport(isResource, schemaPackageFmt)
	values, err := genAttributeValues(isResource, item)
	if err != nil {
		return nil, err
	}

	nested, err := genAttributes(isResource, item.Items, nil)
	if err != nil {
		return nil, err
	}
	values["NestedObject"] = jen.Qual(pkg, "NestedAttributeObject").Values(nested)
	return jen.Qual(pkg, "MapNestedAttribute").Values(dictFromMap(values, false)), nil
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
	if item.IsObject() || item.IsList() {
		t = "ListNestedBlock"
	}

	values["NestedObject"] = jen.Qual(pkg, "NestedBlockObject").Values(nested)
	return jen.Qual(pkg, t).Values(dictFromMap(values, false)), nil
}

func genAttributeValues(isResource bool, item *Item) (map[string]jen.Code, error) {
	values := make(map[string]jen.Code)
	description := fmtDescription(isResource, item)
	if description != "" {
		values["MarkdownDescription"] = jen.Lit(description)
	}

	if item.DeprecationMessage != "" {
		values["DeprecationMessage"] = jen.Lit(item.DeprecationMessage)
	}

	if item.Sensitive {
		values["Sensitive"] = jen.True()
	}

	if !item.IsNested() {
		if isResource {
			typedPlanmodifier := getTypedImport(item.Type, planmodifierTypedImport)
			planModifiers := make([]jen.Code, 0)
			if item.ForceNew {
				planModifiers = append(planModifiers, jen.Qual(typedPlanmodifier, "RequiresReplace").Call())
			}

			if item.UseStateForUnknown {
				planModifiers = append(planModifiers, jen.Qual(typedPlanmodifier, "UseStateForUnknown").Call())
			}

			if len(planModifiers) > 0 {
				values["PlanModifiers"] = jen.Index().Qual(entityImport(isResource, planmodifierPackageFmt), item.TFType()).
					Values(planModifiers...)
			}

			if item.Required {
				values["Required"] = jen.True()
			}

			if item.Optional {
				values["Optional"] = jen.True()
			}

			if item.Computed {
				values["Computed"] = jen.True()
			}
		} else {
			if item.IDAttribute && !item.Virtual {
				values["Required"] = jen.True()
			} else {
				values["Computed"] = jen.True()
			}
		}
	}

	// So far no validations for datasources
	if !item.IsReadOnly(isResource) {
		validators, err := genValidators(item)
		if err != nil {
			return nil, err
		}

		if len(validators) > 0 {
			values["Validators"] = jen.Index().Qual(validatorPackage, item.TFType()).Values(validators...)
		}

		if item.Default != nil {
			values["Default"] = jen.Qual(getTypedImport(item.Type, defaultsTypedImport), "Static"+item.TFType()).Call(jen.Lit(item.Default))
		}
	}

	return values, nil
}
