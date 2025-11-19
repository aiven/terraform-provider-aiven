package main

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"sort"

	"github.com/dave/jennifer/jen"
	"github.com/hashicorp/go-multierror"
	"github.com/iancoleman/strcase"
)

const (
	avnGenHandlerPackage = "github.com/aiven/go-client-codegen/handler/"
	viewSuffix           = "View"
	optionsSuffix        = "Options"
)

// genViews generates CRUD views for the resource, skips disabled or undefined operations.
func genViews(item *Item, def *Definition) ([]jen.Code, error) {
	codes := make([]jen.Code, 0)
	for _, op := range listOperations() {
		v, err := genGenericView(item, def, op)
		if err != nil {
			return nil, fmt.Errorf("%s view error: %w", op, err)
		}

		if v != nil {
			codes = append(codes, v)
		}
	}

	builders := make([]jen.Code, 0)
	if def.Resource != nil {
		builders = append(builders, genNewResource(resourceType, def))
	}

	if def.Datasource != nil {
		builders = append(builders, genNewResource(datasourceType, def))
	}

	return append(builders, codes...), nil
}

// genNewResource generates NewResource or NewDatasource function
func genNewResource(entity entityType, def *Definition) jen.Code {
	// Resource might have multiple read operations.
	// We use here a dict, so we don't need to handle duplicates.
	values := map[string]jen.Code{
		"Beta":     jen.Lit(def.Beta),
		"TypeName": jen.Id(typeName),
		"Schema":   jen.Id(string(entity) + schemaSuffix),
	}

	for _, v := range def.Operations {
		if v.DisableView || entity == datasourceType && v.Type != OperationRead {
			// Datasource only supports Read operation
			continue
		}
		values[firstUpper(v.Type)] = jen.Id(fmt.Sprintf("%s%s", v.Type, viewSuffix))
	}

	entityName := string(entity)
	if entity == resourceType {
		values["IDFields"] = jen.Id(funcIDFields).Call()
		values["RefreshState"] = jen.Lit(def.RefreshState)
	}

	valuesDict := make(jen.Dict)
	for k, v := range values {
		valuesDict[jen.Id(k)] = v
	}

	title := entity.Title() + optionsSuffix
	genericType := jen.List(jen.Op("*").Id(entityName+"Model"), jen.Id(tfRootModel))
	returnValue := jen.Qual(adapterPackage, title).Index(genericType).Values(valuesDict)
	return jen.Var().Id(title).Op("=").Add(returnValue)
}

func genGenericView(item *Item, def *Definition, operation OperationType) (jen.Code, error) {
	operations := make([]Operation, 0)
	for _, op := range def.Operations {
		if op.Type == operation && !op.DisableView {
			operations = append(operations, op)
		}
	}

	if len(operations) == 0 {
		return nil, nil
	}

	view := jen.Func().Id(fmt.Sprintf("%s%s", operation, viewSuffix))
	view.ParamsFunc(func(g *jen.Group) {
		g.Id("ctx").Qual("context", "Context")
		g.Id("client").Qual(avnGenPackage, "Client")

		// Generates the view function.
		// See ResourceOptions for parameter details.
		switch operation {
		case OperationCreate:
			g.Id("plan").Op("*").Id(tfRootModel)
		case OperationUpdate:
			g.List(jen.Id("plan"), jen.Id("state"), jen.Id("config")).Op("*").Id(tfRootModel)
		default:
			g.Id("state").Op("*").Id(tfRootModel)
		}
	})

	// The return value
	view.Qual(diagPackage, "Diagnostics")

	// The function block
	errM := new(multierror.Error)
	view.BlockFunc(func(g *jen.Group) {
		g.Var().Id("diags").Qual(diagPackage, "Diagnostics")
		for i, op := range operations {
			if i != 0 {
				g.If(jen.Id("diags").Dot("HasError").Call()).Block(jen.Return().Id("diags"))
			}

			g.Func().Params().BlockFunc(func(f *jen.Group) {
				err := genGenericViewOperation(f, item, def, op)
				if err != nil {
					errM = multierror.Append(errM, fmt.Errorf("%s view error: %w", op.ID, err))
				}
			}).Call()
		}

		g.Return().Id("diags")
	})

	return view, errM.ErrorOrNil()
}

func genGenericViewOperation(g *jen.Group, item *Item, def *Definition, operation Operation) error {
	inPath := def.Operations.AppearsInID(operation.ID, PathParameter)
	inRequest := def.Operations.AppearsInID(operation.ID, RequestBody)
	inResponse := def.Operations.AppearsInID(operation.ID, ResponseBody)

	// Properties must be sorted for consistent output
	properties := slices.Collect(maps.Values(item.Properties))
	sort.Slice(properties, func(i, j int) bool {
		return properties[i].Name < properties[j].Name
	})

	hasRequest := len(filterAppearsIn(properties, inRequest)) > 0
	hasResponse := len(filterAppearsIn(properties, inResponse)) > 0

	// Generates the code that converts the TF model into the Client request struct
	if hasRequest {
		state := jen.Id("state")
		if operation.Type == OperationCreate {
			state = jen.Nil()
		}

		// Creates the request var
		pkgName := avnGenHandlerPackage + def.ClientHandler
		g.Var().Id("req").Qual(pkgName, string(operation.ID)+"In")

		// Expands data into the request
		g.Id("diags").Dot("Append").Call(
			jen.Id(expandDataFunc).
				Call(
					jen.Id("ctx"),
					jen.Id("plan"),
					state,
					jen.Id("&req"),
				).
				Op("..."),
		)

		// If HasError, then return
		g.If(jen.Id("diags").Dot("HasError").Call()).Block(jen.Return())
		g.Line()
	}

	// Generates the Client request block
	// If there is a response for Delete operation, then renders "_, err :="
	// otherwise "rsp, err :="
	clientCall := jen.Err()
	switch {
	case hasResponse && operation.Type == OperationDelete:
		clientCall = jen.List(jen.Id("_"), jen.Err())
	case hasResponse:
		clientCall = jen.List(jen.Id("rsp"), jen.Err())
	}

	g.Add(clientCall.Op(":=").
		Id("client").Dot(string(operation.ID)).
		CallFunc(func(params *jen.Group) {
			// Read, Update, Delete use the "state"
			state := "state"
			if operation.Type == OperationCreate {
				state = "plan"
			}

			idFields := filterAppearsIn(properties, inPath)
			sort.Slice(idFields, func(i, j int) bool {
				return idFields[i].IDAttributePosition < idFields[j].IDAttributePosition
			})

			// Gathers parameters: ctx and path parameters to call the client method
			params.Id("ctx")
			for _, v := range idFields {
				params.Id(state).Dot(v.GoFieldName()).Dot("Value" + v.TFType()).Call()
			}

			if hasRequest {
				params.Id("&req")
			}
		}),
	)

	// If the client returned the error
	g.If(jen.Err().Op("!=").Nil()).
		Block(
			jen.Id("diags").Dot("Append").Call(
				jen.Qual(errMsgPackage, "FromError").Call(
					jen.Lit(fmt.Sprintf(`%s Error`, firstUpper(operation.ID))),
					jen.Err(),
				),
			),
			jen.Return(),
		)

	// Reads the response from the previous call
	if hasResponse && operation.Type != OperationDelete {
		err := viewResponse(g, item, def, operation)
		if err != nil {
			return err
		}
	}

	return nil
}

func viewResponse(g *jen.Group, item *Item, def *Definition, operation Operation) error {
	state := "plan"
	if operation.Type == OperationRead {
		state = "state"
	}

	// Flatten depends on whether the result is a list or a single object
	flatten := func(rsp jen.Code) jen.Code {
		return jen.Id("diags").Dot("Append").Call(
			jen.Id(flattenDataFunc).
				Call(
					jen.Id("ctx"),
					jen.Id(state),
					rsp,
				).
				Op("..."),
		)
	}

	// Wraps the list into a map by the specified key
	if operation.ResultListToKey != "" {
		m := jen.Op("&").Map(jen.String()).Any()
		g.Add(flatten(m.Values(jen.Dict{jen.Lit(operation.ResultListToKey): jen.Id("rsp")})))
		return nil
	}

	// Flattens the response directly, because it is not a list but an object
	if len(operation.ResultListLookupKeys) == 0 {
		g.Add(flatten(jen.Id("rsp")))
		return nil
	}

	// Finds a match in the list.
	// E.g. if rsp[0].foo == state.Foo && rsp[0].bar == state.Bar ...
	fieldsMatch := make([]jen.Code, 0)
	for i, key := range operation.ResultListLookupKeys {
		if i > 0 {
			fieldsMatch = append(fieldsMatch, jen.Op("&&"))
		}

		field := item.Properties[key]
		fieldsMatch = append(
			fieldsMatch,
			jen.Id("v").Dot(clientCamelName(key)).
				Op("==").
				Id(state).Dot(field.GoFieldName()).Dot("Value"+field.TFType()).Call(),
		)
	}

	// Flattens the item on the list if all key fields match.
	g.For(jen.List(jen.Id("_"), jen.Id("v")).Op(":=").Range().Id("rsp")).
		Block(jen.If(fieldsMatch...).Block(flatten(jen.Id("&v")), jen.Return()))

	// If passed the "for" loop, then no item was found - adds error
	g.Id("diags").Dot("AddError").Call(
		jen.Lit("Resource Not Found"),
		jen.Lit(fmt.Sprintf("`%s` with given %s not found", def.typeName, humanizeCodeList(operation.ResultListLookupKeys))),
	)

	return nil
}

// filterAppearsIn filters items by their appearance in Create/Update request, response, etc.
func filterAppearsIn(items []*Item, appearsIn AppearsIn) []*Item {
	var props []*Item
	for _, v := range items {
		if v.AppearsIn.Contains(appearsIn) {
			props = append(props, v)
		}
	}
	return props
}

// humanizeCodeList turns ["foo", "bar", "baz"] -> "`foo`, `bar` and `baz`"
func humanizeCodeList(args []string) string {
	list := ""
	last := len(args) - 1
	for i, v := range args {
		switch i {
		case 0:
		case last:
			list += " and "
		default:
			list += ", "
		}
		list += fmt.Sprintf("`%s`", v)
	}
	return list
}

var reNonWord = regexp.MustCompile(`\W+`)

// clientCamelName returns go client camel case name
// fixme: make client naming same as in terraform
func clientCamelName(s string) string {
	return strcase.ToCamel(reNonWord.ReplaceAllString(s, "_"))
}
