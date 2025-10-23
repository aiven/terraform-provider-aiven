package main

import (
	"path/filepath"
	"strings"

	"github.com/dave/jennifer/jen"
)

func genModels(item *Item) ([]jen.Code, error) {
	models := collectModels(item)
	dataModels := make([]jen.Code, 0)
	dtoModels := make([]jen.Code, 0)
	for _, v := range models {
		data, err := genItemTFModel(v)
		if err != nil {
			return nil, err
		}
		dataModels = append(dataModels, data)
		if v.IsRoot() {
			dataModels = append(dataModels, genSetIDMeth(item))
		}

		dto, err := genItemApiModel(v)
		if err != nil {
			return nil, err
		}
		dtoModels = append(dtoModels, dto)
	}
	return append(dataModels, dtoModels...), nil
}

// genTFModel generates a specific TF model for resource and datasource (prefixed)
func genTFModel(isResource bool, item *Item) []jen.Code {
	structFields := []jen.Code{
		// dataModel
		jen.Id(item.TFModelName()),

		// Timeouts
		jen.Id("Timeouts").
			Qual(entityImport(isResource, timeoutsPackageFmt), "Value").
			Tag(genFieldTag("tfsdk", "timeouts")),
	}

	entity := boolEntity(isResource)
	name := string(entity) + "Model"
	meth := jen.Func().Params(jen.Id(tfVar).Op("*").Id(name))
	return []jen.Code{
		// Generates the model
		jen.
			Commentf("%s with specific %s timeouts", name, entity).Line().
			Type().Id(name).
			Struct(structFields...).
			Line(),

		// Adds SharedModel method that returns *tfModel
		meth.Clone().
			Id("SharedModel").Params().Op("*").Id(item.TFModelName()).
			Block(jen.Return(jen.Op("&").Id(tfVar).Dot(item.TFModelName()))).
			Line().
			Line(), // Had to add this extra to separate methods visually

		// Adds TimeoutsObject method that returns types.Object (see TimeoutsObject usage)
		meth.Clone().
			Id("TimeoutsObject").Params().Qual(typesPackage, "Object").
			Block(jen.Return(jen.Id(tfVar).Dot("Timeouts").Dot("Object"))).
			Line(),
	}
}

// collectModels collects Item which must have a data model
func collectModels(item *Item) []*Item {
	if item.IsScalar() || item.Items != nil && item.Items.IsScalar() {
		// Scalars and maps/arrays with scalar items do not need models.
		return nil
	}

	items := make([]*Item, 0)
	if item.Items != nil {
		// Array of objects or a map of objects
		items = append(items, collectModels(item.Items)...)
	} else {
		// Object properties
		items = append(items, item)
		for _, k := range sortedKeys(item.Properties) {
			items = append(items, collectModels(item.Properties[k])...)
		}
	}
	return items
}

// genItemTFModel generates TF model and nested models
func genItemTFModel(item *Item) (jen.Code, error) {
	fields := make([]jen.Code, 0)

	if item.IsRoot() {
		// Adds ID field
		fields = append(
			fields,
			jen.Id("ID").Qual(typesPackage, "String").Tag(genFieldTag("tfsdk", "id")),
		)
	}

	for _, k := range sortedKeys(item.Properties) {
		if k == "id" && item.IsRoot() {
			// The id field is already added
			continue
		}

		v := item.Properties[k]
		fields = append(
			fields,
			jen.Id(v.GoFieldName()).Qual(typesPackage, v.TFType()).Tag(genFieldTag("tfsdk", v.Name)),
		)
	}

	modelName := item.TFModelName()
	model := jen.Type().Id(modelName).Struct(fields...)
	return model, nil
}

// genItemApiModel generates TF model and nested models
func genItemApiModel(item *Item) (jen.Code, error) {
	fields := make([]jen.Code, 0)

	for _, k := range sortedKeys(item.Properties) {
		v := item.Properties[k]
		// It is important to use JSONName here, it might be different from Name
		tags := genFieldTag("json", v.JSONName)
		if !v.Nullable {
			tags["json"] += ",omitempty"
		}

		// All fields nullable to make sure omitempty won't omit empty slices/maps
		t := v.ApiType()
		if !strings.HasPrefix(t, "*") {
			t = "*" + t
		}
		fields = append(
			fields,
			jen.Id(v.GoFieldName()).Id(t).Tag(tags),
		)
	}

	model := jen.Type().Id(item.ApiModelName()).Struct(fields...)
	return model, nil
}

func genFieldTag(kv ...string) map[string]string {
	m := make(map[string]string)
	for i := 0; i < len(kv); i += 2 {
		m[kv[i]] = kv[i+1]
	}
	return m
}

const funcIDFields = "idFields"

func genIDFields(avnName string, fields []string) jen.Code {
	codes := make([]jen.Code, 0)
	for _, v := range fields {
		codes = append(codes, jen.Lit(v))
	}

	return jen.
		Commentf("%s the ID attribute fields, i.e.:", funcIDFields).Line().
		Commentf("terraform import %s.foo %s", avnName, strings.ToUpper(filepath.Join(fields...))).Line().
		Func().Id(funcIDFields).Params().Index().String().
		Block(jen.Return().Index().String().Values(codes...))
}

func genSetIDMeth(item *Item) jen.Code {
	idFields := item.GetIDFields()
	block := make([]jen.Code, 0)
	params := make([]jen.Code, 0)
	pathValues := make([]jen.Code, 0)

	for _, v := range idFields {
		// Func params
		params = append(params, jen.Id(v.GoVarName()).String())

		// Path values
		pathValues = append(pathValues, jen.Id(v.GoVarName()))

		// Fields to set
		if v.Name != "id" {
			setValues := jen.Qual(typesPackage, v.TFType()+"Value").Call(jen.Id(v.GoVarName()))
			block = append(block, jen.Id(tfVar).Dot(v.GoFieldName()).Op("=").Add(setValues))
		}
	}

	block = append(
		block,
		jen.Id(tfVar).Dot("ID").Op("=").
			Qual(typesPackage, "StringValue").
			Call(jen.Qual("path/filepath", "Join").Call(pathValues...)),
	)

	return jen.Func().
		Parens(jen.Id(tfVar).Op("*").Id(item.TFModelName())).
		Id(setIDFunc).
		Params(params...).
		Block(block...)
}
