package main

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/dave/jennifer/jen"
)

const (
	flattenDataFunc = "flattenData"
	expandDataFunc  = "expandData"
	attrPrefix      = "attrs"
	planVar         = "plan"
	stateVar        = "state"
	tfVar           = "tf"
	tfModelPrefix   = "tfModel"
	apiVar          = "api"
	apiModelPrefix  = "apiModel"
	setIDFunc       = "SetID"
	// To avoid conflicts with generated models, the root models must have unique ever name.
	tfRootModel  = tfModelPrefix
	apiRootModel = apiModelPrefix
)

// genFlatten generates a function that turns rsp into TF object
func genFlatten(item *Item) ([]jen.Code, error) {
	block := []jen.Code{
		jen.Id(apiVar).Op(":=").New(jen.Id(item.ApiModelName())),

		// Renders Unmarshal
		jen.Id("err").
			Op(":=").
			Qual(utilPackage, "Remarshal").
			Call(
				jen.Id("rsp"),
				jen.Id(apiVar),
				jen.Id("modifiers").Op("..."),
			),

		jen.Add(ifErr(true, "Remarshal error", "Failed to remarshal Response to dtoModel: %s")),
	}

	props, err := genFlattenProperties(item, true)
	if err != nil {
		return nil, err
	}

	block = append(block, props...)

	// Often ID field values are not set in the Response.
	// But they are available in the ID field.
	// Generates the code that splits the ID field and sets the values to the fields.
	// This way the plan command won't output a diff.
	if _, ok := item.Properties["id"]; !ok {
		block = append(
			block,
			genSetIDFields(item),
			genSetID(item),
		)
	}

	block = append(
		block,
		jen.Return(jen.Nil()),
	)

	flatten := jen.
		Comment(flattenDataFunc+" turns Response into TF object").Line().
		Func().
		Id(flattenDataFunc).
		Index(jen.Id("R").Any()).
		Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id(stateVar).Op("*").Id(item.TFModelName()),
			jen.Id("rsp").Id("*R"),
			jen.Id("modifiers").Op("...").Qual(utilPackage, "MapModifier").Index(jen.Id("R")),
		).
		Qual(diagPackage, "Diagnostics").
		Block(block...)

	codes := []jen.Code{flatten}
	attrs := make([]jen.Code, 0)
	for i, v := range collectModels(item) {
		if i == 0 {
			// the root is already in "codes"
			continue
		}
		code, err := genFlattenModel(v)
		if err != nil {
			return nil, err
		}

		codes = append(codes, code)
		attrs = append(attrs, genAttrFunc(v))
	}

	return append(codes, attrs...), nil
}

func genFlattenAttribute(item *Item, rootLevel bool) (*jen.Statement, error) {
	dtoField := jen.Id(apiVar).Dot(item.GoFieldName())
	value := jen.Id(stateVar).Dot(item.GoFieldName()).Op("=")
	block := make([]jen.Code, 0)
	notNil := dtoField.Clone().Op("!=").Nil()
	switch {
	case item.IsScalar():
		value.Qual(typesPackage, item.TFType()+"PointerValue").Call(dtoField.Clone())
		if item.IsRootProperty() && !item.Nullable && item.Type == SchemaTypeString {
			// Adds (*dto.State != "" || !state.State.IsNull())
			// Ignores empty strings if user hasn't explicitly set it in the state.
			notNil.Op("&&").Parens(
				jen.Op("*").Id(apiVar).Dot(item.GoFieldName()).Op("!=").Lit("").Op("||").
					Op("!").Id(stateVar).Dot(item.GoFieldName()).Dot("IsNull").Call(),
			)
		}
	case item.IsNested():
		// A struct or an array of structs.
		value.Id(item.GoVarName())
		val := jen.List(jen.Id(item.GoVarName()), jen.Id("diags")).Op(":=")
		if item.IsObject() {
			val.Qual(utilPackage, "FlattenSingleNested").
				Call(
					jen.Id("ctx"),
					jen.Id("flatten"+item.UniqueName()),
					jen.Add(dtoField.Clone()),
					jen.Id(attrPrefix+item.UniqueName()).Call(),
				)
		} else {
			val.Qual(utilPackage, "FlattenSetNested").
				Call(
					jen.Id("ctx"),
					jen.Id("flatten"+item.UniqueName()),
					jen.Add(jen.Op("*").Add(dtoField.Clone())),
					jen.Id(attrPrefix+item.UniqueName()).Call(),
				)
		}
		block = append(block, val, ifHasError(rootLevel))
	case item.IsArray(), item.IsMap():
		// 1. If the API returns an empty list/map, while the field is nil, TF will output an error.
		//    Checks the length of the list/map to not set empty array/map.
		// 2. But if we have sent the empty array to the API, it is OK to get it back — checks the state for nil.
		if item.Parent.IsRoot() {
			notNil.Op("&&").Parens(
				jen.Len(jen.Op("*").Add(dtoField.Clone())).Op(">").Lit(0).
					Op("||").Op("!").Id(stateVar).Dot(item.GoFieldName()).Dot("IsNull").Call(),
			)
		}

		value.Id(item.GoVarName())
		switch {
		case item.Items.IsScalar():
			// A list or a map of scalars.
			val := jen.List(jen.Id(item.GoVarName()), jen.Id("diags")).Op(":=").
				Qual(typesPackage, item.TFType()+"ValueFrom").
				Call(
					jen.Id("ctx"),
					jen.Qual(typesPackage, item.Items.TFType()+"Type"),
					dtoField.Clone(),
				)
			block = append(block, val, ifHasError(rootLevel))
		case item.IsMapNested():
			// This is a map with non-scalar values
			val := jen.List(jen.Id(item.GoVarName()), jen.Id("diags")).Op(":=")
			val.Qual(utilPackage, "FlattenMapNested").
				Call(
					jen.Id("ctx"),
					jen.Id("flatten"+item.UniqueName()),
					jen.Op("*").Add(dtoField.Clone()),
					jen.Id(attrPrefix+item.UniqueName()).Call(),
				)
			block = append(block, val, ifHasError(rootLevel))
		default:
			return nil, fmt.Errorf("unsupported items type map/array %s for %s", item.Items.Type, item.Path())
		}
	default:
		return nil, fmt.Errorf("unsupported type %s for %s", item.Type, item.Path())
	}

	return jen.If(notNil).Block(append(block, value)...), nil
}

func genFlattenProperties(item *Item, rootLevel bool) ([]jen.Code, error) {
	block := make([]jen.Code, 0, len(item.Properties))
	for _, k := range nestedFirst(item.Properties) {
		value, err := genFlattenAttribute(item.Properties[k], rootLevel)
		if err != nil {
			return nil, err
		}
		block = append(block, value)
	}
	return block, nil
}

func genSetID(item *Item) *jen.Statement {
	values := make([]jen.Code, 0)
	for _, v := range item.GetIDFields() {
		values = append(values, jen.Id(stateVar).Dot(v.GoFieldName()).Dot("Value"+v.TFType()).Call())
	}

	return jen.Id(stateVar).Dot(setIDFunc).Call(values...)
}

func genSetIDFields(item *Item) jen.Code {
	idFields := item.GetIDFields()

	// Splits ID into parts.
	idValue := jen.Id(stateVar).Dot("ID").Dot("ValueString").Call()
	length := len(idFields)
	codes := []jen.Code{
		jen.Var().Id("parts").Index(jen.Lit(length)).String(),
		jen.For(
			jen.List(jen.Id("i"), jen.Id("v")).
				Op(":=").Range().Qual("strings", "SplitN").
				Call(idValue.Clone(), jen.Lit("/"), jen.Lit(length)).
				Block(jen.Id("parts").Index(jen.Id("i")).Op("=").Id("v")),
		),
	}

	// Sets parts to the ID fields.
	for i, v := range idFields {
		field := jen.Id(stateVar).Dot(v.GoFieldName())
		empty := field.Clone().Dot("ValueString").Call().Op("==").Lit("")
		value := jen.Qual(typesPackage, "StringValue").Call(jen.Id("parts").Index(jen.Lit(i)))
		codes = append(
			codes,
			jen.If(empty).Block(field.Clone().Op("=").Add(value)),
		)
	}

	return jen.
		Comment("Response may not contain ID fields.").Line().
		Comment("In that case, `terraform import` won't be able to set them. Gets values from the ID.").Line().
		If(idValue.Clone().Op("!=").Lit("")).Block(codes...)
}

func genFlattenModel(item *Item) (jen.Code, error) {
	props, err := genFlattenProperties(item, false)
	if err != nil {
		return nil, err
	}

	f := jen.Func().
		Id("flatten"+item.UniqueName()).
		Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id(apiVar).Id(item.ApiType()),
		).
		Parens(jen.List(
			jen.Op("*").Id(item.TFModelName()),
			jen.Qual(diagPackage, "Diagnostics"),
		)).
		BlockFunc(func(g *jen.Group) {
			g.Add(jen.Id(stateVar).Op(":=").New(jen.Id(item.TFModelName())))
			for _, p := range props {
				g.Add(p)
			}
			g.Add(jen.Return(jen.Id(stateVar), jen.Nil()))
		})

	return f, nil
}

// nestedFirst returns Nested Items first to make code visually easier to read.
func nestedFirst(m map[string]*Item) []string {
	keys := slices.Collect(maps.Keys(m))
	sort.Slice(keys, func(i, j int) bool {
		a := m[keys[i]].IsScalar()
		b := m[keys[j]].IsScalar()
		if a == b {
			return strings.Compare(m[keys[i]].Name, m[keys[j]].Name) < 0
		}
		return b
	})
	return keys
}

func genAttrFunc(item *Item) *jen.Statement {
	return jen.Func().
		Id(attrPrefix+item.UniqueName()).
		Params().Qual(typesPackage, "ObjectType").
		BlockFunc(func(g *jen.Group) {
			values := make(jen.Dict)
			for k, v := range item.Properties {
				values[jen.Lit(k)] = genAttrFieldType(v)
			}

			attrTypes := jen.Map(jen.String()).Qual(attrPackage, "Type").Values(values)
			g.Return(
				jen.Qual(typesPackage, "ObjectType").Values(jen.Dict{
					jen.Id("AttrTypes"): attrTypes,
				}),
			)
		})
}

func genAttrFieldType(item *Item) *jen.Statement {
	switch {
	case item.IsScalar(), item.IsMap():
		return jen.Qual(typesPackage, item.TFType()+"Type")
	case item.IsObject():
		return jen.Id(attrPrefix + item.UniqueName()).Call()
	case item.IsArray():
		return jen.Qual(typesPackage, "SetType").Values(jen.Dict{
			jen.Id("ElemType"): genAttrFieldType(item.Items),
		})
	default:
		panic(fmt.Sprintf("genAttrFieldType, unknown item type: %s", item.TFType()))
	}
}

func ifErr(rootLevel bool, summary, format string) jen.Code {
	return jen.
		If(jen.Err().Op("!=").Nil()).
		BlockFunc(func(g *jen.Group) {
			g.Var().Id("diags").Qual(diagPackage, "Diagnostics")
			g.Id("diags").Dot("AddError").Call(
				jen.Lit(summary),
				jen.Qual("fmt", "Sprintf").Call(jen.Lit(format), jen.Err().Dot("Error").Call()),
			)
			g.Add(returnErr(rootLevel))
		})
}

func ifHasError(rootLevel bool) jen.Code {
	return jen.If(jen.Id("diags").Dot("HasError").Call()).Block(returnErr(rootLevel))
}

func returnErr(rootLevel bool) jen.Code {
	if rootLevel {
		return jen.Return(jen.Id("diags"))
	}
	return jen.Return(jen.Nil(), jen.Id("diags"))
}
