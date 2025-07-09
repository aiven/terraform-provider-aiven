package main

import (
	"fmt"

	"github.com/dave/jennifer/jen"
)

// genExpand generates a function that turns TF object into DTO
func genExpand(item *Item) ([]jen.Code, error) {
	props, err := genExpandModelProperties(item, true)
	if err != nil {
		return nil, err
	}

	var block []jen.Code
	block = append(block, props...)
	block = append(
		block,
		jen.Id("err").
			Op(":=").
			Qual(utilPackage, "Unmarshal").
			Call(
				jen.Id(apiVar),
				jen.Id("req"),
				jen.Id("modifiers").Op("..."),
			),
		jen.Add(ifErr(true, "Unmarshal error", "Failed to unmarshal dtoModel to Request: %s")),
		jen.Return(jen.Nil()),
	)

	expand := jen.
		Comment("expandData turns TF object into Request").Line().
		Func().
		Id("expandData").
		Index(jen.Id("R").Any()).
		Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.List(jen.Id(planVar), jen.Id(stateVar)).Op("*").Id(tfRootModel),
			jen.Id("req").Id("*R"),
			jen.Id("modifiers").Op("...").Qual(utilPackage, "MapModifier").Index(jen.Id(apiRootModel)),
		).
		Qual(diagPackage, "Diagnostics").
		Block(block...)

	codes := []jen.Code{expand}
	for i, v := range collectModels(item) {
		if i == 0 {
			// the root is already in "codes"
			continue
		}
		code, err := genExpandModel(v)
		if err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}

	return codes, nil
}

func genExpandModelProperties(item *Item, rootLevel bool) ([]jen.Code, error) {
	block := []jen.Code{
		jen.Id(apiVar).Op(":=").New(jen.Id(item.ApiModelName())),
	}

	for _, k := range nestedFirst(item.Properties) {
		v := item.Properties[k]
		if !v.IsReadOnly(true) {
			value, err := genExpandField(v, rootLevel)
			if err != nil {
				return nil, err
			}
			block = append(block, value)
		}
	}
	return block, nil
}

func genExpandField(item *Item, rootLevel bool) (*jen.Statement, error) {
	dataField := jen.Id(planVar).Dot(item.GoFieldName())
	value := jen.Id(apiVar).Dot(item.GoFieldName()).Op("=")
	block := make([]jen.Code, 0)
	switch {
	case item.IsScalar():
		if item.Nullable {
			value.Id(planVar).Dot(item.GoFieldName()).Dot("Value" + item.TFType() + "Pointer").Call()
			block = append(block, value)
		} else {
			val := jen.Id(item.GoVarName()).Op(":=").Id(planVar).Dot(item.GoFieldName()).Dot("Value" + item.TFType()).Call()
			value.Op("&").Id(item.GoVarName())
			block = append(block, val, value)
		}
	case item.IsNested():
		f := "ExpandSingleNested"
		if item.IsArray() {
			f = "ExpandSetNested"
			value.Op("&")
		}
		value.Id(item.GoVarName())

		val := jen.List(jen.Id(item.GoVarName()), jen.Id("diags")).Op(":=")
		val.Qual(utilPackage, f)
		val.Call(
			jen.Id("ctx"),
			jen.Id("expand"+item.UniqueName()),
			dataField.Clone(),
		)
		block = append(block, val, ifHasError(rootLevel), value)
	case item.IsArray(), item.IsMap():
		if !item.Items.IsScalar() {
			return nil, fmt.Errorf("unsupported type %s for %s", item.Items.Type, item.Path())
		}

		varVal := jen.Id(item.GoVarName()).Op(":=").Make(jen.Id(item.ApiType()), jen.Lit(0))
		val := jen.Id("diags").Op(":=").
			Add(dataField.Clone()).
			Dot("ElementsAs").
			Call(
				jen.Id("ctx"),
				jen.Op("&").Id(item.GoVarName()),
				jen.False(),
			)
		value.Op("&").Id(item.GoVarName())
		block = append(block, varVal, val, ifHasError(rootLevel), value)
	default:
		return nil, fmt.Errorf("unsupported type %s for %s", item.Type, item.Path())
	}

	ifOk := jen.Op("!").Add(dataField.Clone()).Dot("IsNull").Call()
	if item.Computed {
		// Computed fields have initially state "unknown", not "null".
		ifOk.Op("&&").Op(" !").Add(dataField.Clone()).Dot("IsUnknown").Call()
	} else if item.IsRootProperty() {
		// Compares root fields with state.
		ifOk.Op("||").Id(stateVar).Op("!=").Nil().
			Op("&&").Op("!").Id(stateVar).Dot(item.GoFieldName()).Dot("IsNull").Call()
	}

	return jen.If(ifOk).Block(block...), nil
}

// genExpandModel for instance, expandModelAcls()
func genExpandModel(item *Item) (jen.Code, error) {
	props, err := genExpandModelProperties(item, false)
	if err != nil {
		return nil, err
	}

	function := jen.Func().
		Id("expand"+item.UniqueName()).
		Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id(planVar).Op("*").Id(item.TFModelName()),
		).
		Parens(jen.List(
			jen.Id(item.ApiType()),
			jen.Qual(diagPackage, "Diagnostics"),
		)).
		Block(append(props, jen.Return(jen.Id(apiVar), jen.Nil()))...)

	return function, nil
}
