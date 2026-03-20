package main

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/hashicorp/go-multierror"
	"github.com/samber/lo"
)

const (
	viewSuffix                 = "View"
	optionsSuffix              = "Options"
	configValidatorsFuncSuffix = "ConfigValidators"
	planModifier               = "planModifier"
	resourceData               = "ResourceData"
	renameFieldsModifier       = "RenameFields"
	flattenModifier            = "flattenModifier"
	expandModifier             = "expandModifier"
	refreshStateWaiter         = "refreshStateWaiter"
)

// genViews generates CRUD views for the resource, skips disabled or undefined operations.
func genViews(item *Item, def *Definition) ([]jen.Code, error) {
	codes := make([]jen.Code, 0)
	for _, op := range listOperationTypes() {
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
		builders = append(builders, genNewResource(resourceType, def, false))
	}

	if def.Datasource != nil {
		hasValidators := len(def.Datasource.ExactlyOneOf) > 0
		if hasValidators {
			codes = append(codes, datasourceExactlyOneOf(def))
		}
		builders = append(builders, genNewResource(datasourceType, def, hasValidators))
	}

	return append(builders, codes...), nil
}

// genNewResource generates NewResource or NewDatasource function
func genNewResource(entity entityType, def *Definition, hasConfigValidators bool) jen.Code {
	// Resource might have multiple read operations.
	// We use here a dict, so we don't need to handle duplicates.
	values := map[string]jen.Code{
		"TypeName":       jen.Id(typeName),
		"Schema":         jen.Id(string(entity) + schemaSuffix),
		"SchemaInternal": jen.Id(fmt.Sprintf(schemaInternalFmt, entity)).Call(),
		"IDFields":       jen.Id(funcIDFields).Call(),
	}

	if lo.FromPtr(def.Beta) {
		values["Beta"] = jen.True()
	}

	for _, v := range def.Operations {
		if entity == datasourceType && v.Type != OperationRead {
			// Datasource only supports Read operation
			continue
		}
		values[firstUpper(v.Type)] = jen.Id(fmt.Sprintf("%s%s", v.Type, viewSuffix))
	}

	if entity.isResource() {
		if def.Resource.RefreshState {
			values["RefreshState"] = jen.True()

			if def.Resource.RefreshStateDelay != 0 {
				values["RefreshStateDelay"] = jen.Qual(adapterPackage, "MustParseDuration").Call(jen.Lit(def.Resource.RefreshStateDelay.String()))
			}

			if def.Resource.RefreshStateWaiter {
				values["RefreshStateWaiter"] = jen.Id(refreshStateWaiter)
			}
		}

		if def.Resource.RemoveMissing {
			values["RemoveMissing"] = jen.True()
		}

		if def.Resource.TerminationProtection {
			values["TerminationProtection"] = jen.True()
		}
	}

	if hasConfigValidators {
		values["ConfigValidators"] = jen.Id(entity.String() + configValidatorsFuncSuffix)
	}

	title := entity.Title() + optionsSuffix
	returnValue := jen.Qual(adapterPackage, title).Values(dictFromMap(values, false))
	return jen.Var().Id(title).Op("=").Add(returnValue)
}

func genGenericView(item *Item, def *Definition, operation OperationType) (jen.Code, error) {
	operations := make([]*Operation, 0)
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
		g.Id("d").Qual(adapterPackage, resourceData)
	})

	// The return value
	view.Error()

	// The function block
	errM := new(multierror.Error)
	view.BlockFunc(func(g *jen.Group) {
		if def.PlanModifier && operation == OperationRead {
			g.Err().Op(":=").Id(planModifier).Call(
				jen.Id("ctx"),
				jen.Id("client"),
				jen.Id("d"),
			)
			g.If(jen.Err().Op("!=").Nil()).Block(jen.Return().Err())
		}

		for i, op := range operations {
			err := genGenericViewOperation(g, i, len(operations), item, def, op)
			if err != nil {
				errM = multierror.Append(errM, fmt.Errorf("%s view error: %w", op.ID, err))
			}
		}
	})

	return view, errM.ErrorOrNil()
}

func genGenericViewOperation(g *jen.Group, funcIndex, funcCount int, item *Item, def *Definition, operation *Operation) error {
	inPath := def.Operations.AppearsInID(operation.ID, operation.Type, PathParameter)
	inRequest := def.Operations.AppearsInID(operation.ID, operation.Type, RequestBody)
	inResponse := def.Operations.AppearsInID(operation.ID, operation.Type, ResponseBody)

	// Properties must be sorted for consistent output
	properties := slices.Collect(maps.Values(item.Properties))
	sort.Slice(properties, func(i, j int) bool {
		return properties[i].Name < properties[j].Name
	})

	hasRequest := len(filterAppearsIn(properties, inRequest)) > 0
	hasResponse := len(filterAppearsIn(properties, inResponse)) > 0

	// Generates the code that converts the TF model into the Client request struct
	if hasRequest {
		// Creates the request var
		pkgName := avnGenHandlerPackage + def.ClientHandler
		g.Id("req").Op(":=").New(jen.Qual(pkgName, string(operation.ID)+"In"))
		g.Err().Op(":=").Id("d").Dot("Expand").CallFunc(func(g *jen.Group) {
			g.Id("req")
			if def.ExpandModifier {
				g.Id(expandModifier).Call(jen.Id("ctx"), jen.Id("client"))
			}
			if len(def.Rename) > 0 && operation.Request != nil {
				renameFields := make(jen.Dict)
				for jsonName, tfName := range def.Rename {
					_, ok := operation.Request.Properties[jsonName]
					if ok {
						renameFields[jen.Lit(tfName)] = jen.Lit(jsonName)
					}
				}

				if len(renameFields) > 0 {
					g.Qual(adapterPackage, renameFieldsModifier).Call(jen.Map(jen.String()).String().Values(renameFields))
				}
			}
		})
		g.If(jen.Err().Op("!=").Nil()).Block(jen.Return().Err())
	}

	// If this is the last function without response, just return the error
	mustReturnResult := funcIndex == funcCount-1 && !hasResponse

	// Generates the Client request block
	// If there is a response for Delete operation, then renders "_, err :="
	// otherwise "rsp, err :="
	var clientCall *jen.Statement
	switch {
	case mustReturnResult:
		clientCall = jen.Return()
	case funcIndex == 0:
		clientCall = jen.Err().Op(":=")
	default:
		clientCall = jen.Err().Op("=")
	}

	rspName := "rsp"
	if funcIndex > 0 {
		rspName = fmt.Sprintf("rsp%d", funcIndex+1)
	}

	switch {
	case operation.Type == OperationDelete && hasResponse:
		clientCall = jen.List(jen.Id("_"), jen.Err()).Op(":=")
	case hasResponse:
		clientCall = jen.List(jen.Id(rspName), jen.Err()).Op(":=")
	}

	g.Add(clientCall.
		Id("client").Dot(string(operation.ID)).
		CallFunc(func(params *jen.Group) {
			idFields := filterAppearsIn(properties, inPath)
			sort.Slice(idFields, func(i, j int) bool {
				return idFields[i].IDAttributePosition < idFields[j].IDAttributePosition
			})

			// Some resources might change the ID fields during Update operation,
			// so we need to use get values from the state instead of the plan.
			method := "Get"
			if operation.Type == OperationUpdate && !item.Properties["id"].UseStateForUnknown {
				method = "GetState"
			}

			// Gathers parameters: ctx and path parameters to call the client method
			params.Id("ctx")
			for _, v := range idFields {
				params.Id("d").Dot(method).Call(jen.Lit(v.Name)).Op(".").Parens(jen.Id(v.GoType()))
			}

			if hasRequest {
				params.Id("req")
			}
		}),
	)

	if mustReturnResult {
		return nil
	}

	if !hasResponse || operation.Type == OperationDelete {
		g.Return().Err()
		return nil
	}

	// Reads the response from the previous call
	g.If(jen.Err().Op("!=").Nil()).Block(jen.Return().Err())
	flatten := func(rsp jen.Code) jen.Code {
		v := jen.Id("d").Dot("Flatten").CallFunc(func(g *jen.Group) {
			g.Add(rsp)

			if len(def.Rename) > 0 && operation.Response != nil {
				renameFields := make(jen.Dict)
				for jsonName, tfName := range def.Rename {
					_, ok := operation.Response.Properties[jsonName]
					if ok {
						renameFields[jen.Lit(jsonName)] = jen.Lit(tfName)
					}
				}

				if len(renameFields) > 0 {
					g.Qual(adapterPackage, renameFieldsModifier).Call(jen.Map(jen.String()).String().Values(renameFields))
				}
			}

			if def.FlattenModifier {
				g.Id(flattenModifier).Call(jen.Id("ctx"), jen.Id("client"))
			}
		})

		if funcCount == 1 || funcIndex == funcCount-1 {
			v = jen.Return(v)
		} else {
			v = jen.Err().Op("=").Add(v).
				Line().
				If(jen.Err().Op("!=").Nil()).Block(jen.Return().Err())
		}
		return v
	}

	// Wraps the result into a map by the specified key
	if operation.ResultToKey != "" {
		m := jen.Op("&").Map(jen.String()).Any()
		g.Add(flatten(m.Values(jen.Dict{jen.Lit(operation.ResultToKey): jen.Id(rspName)})))
		return nil
	}

	// Flattens the response directly, because it is not a list but an object
	if len(operation.ResultListLookupKeys) == 0 {
		g.Add(flatten(jen.Id(rspName)))
		return nil
	}

	// Finds a match in the list.
	// E.g. rsp[0].foo == state.Foo && rsp[0].bar == state.Bar ...
	var fieldsMatch jen.Statement
	for i, key := range sortedKeys(operation.ResultListLookupKeys) {
		fieldName := operation.ResultListLookupKeys[key]
		if i > 0 {
			if operation.ResultListLookupKeysOr {
				fieldsMatch.Op("||")
			} else {
				fieldsMatch.Op("&&")
			}
		}

		field, ok := item.Properties[fieldName]
		if !ok {
			return fmt.Errorf("unknown lookup key %q in result list for operation %q", fieldName, operation.ID)
		}

		fieldsMatch.Id("rsp").Index(jen.Id("i")).Dot(key).
			Op("==").
			Id("d").Dot("Get").Call(jen.Lit(field.Name)).Op(".").Parens(jen.Id(field.GoType()))
	}

	// Filters the response by the fields match
	const foundName = "found"
	g.Id(foundName).Op(":=").Qual(adapterPackage, "FilterIndex").CallFunc(func(g *jen.Group) {
		g.Id(rspName)
		g.Func().Params(jen.Id("i").Int()).Bool().Block(jen.Return(&fieldsMatch))
	})

	// Makes a list of human readable values, e.g. foo, bar and baz
	values := lo.Values(operation.ResultListLookupKeys)
	slices.Sort(values)
	humanValues := humanizeCodeList(values, func() string {
		if operation.ResultListLookupKeysOr {
			return " or "
		}
		return " and "
	}())

	// Switches on the number of found items
	// 0: returns not found error
	// 1: flattens the item
	// 2+: returns multiple found error
	g.Switch(jen.Len(jen.Id(foundName))).Block(
		jen.Case(jen.Lit(1)).Block(flatten(jen.Op("&").Id(foundName).Index(jen.Lit(0)))),
		jen.Case(jen.Lit(0)).BlockFunc(func(g *jen.Group) {
			// If passed the "for" loop, then no item was found - adds error
			g.Return().
				Qual(avnGenPackage, "Error").Values(jen.Dict{
				jen.Id("OperationID"): jen.Lit(string(operation.ID)),
				jen.Id("Status"):      jen.Qual("net/http", "StatusNotFound"),
				jen.Id("Message"):     jen.Lit(fmt.Sprintf("`%s` with given %s not found", def.typeName, humanValues)),
			})
		}),
		jen.Default().Block(jen.Return().Qual("fmt", "Errorf").Call(
			jen.Lit(fmt.Sprintf("found %%d `%s` with given %s", def.typeName, humanValues)),
			jen.Len(jen.Id(foundName)),
		)),
	)

	return nil
}

func datasourceExactlyOneOf(def *Definition) jen.Code {
	fields := make([]jen.Code, len(def.Datasource.ExactlyOneOf))
	for i, v := range def.Datasource.ExactlyOneOf {
		fields[i] = jen.Qual(pathPackage, "MatchRoot").Call(jen.Lit(v))
	}

	pkg := datasourceType.Import(entityValidatorPkgFmt)
	validators := jen.Index().Qual(datasourcePkg, "ConfigValidator").Custom(
		multilineValues(), jen.Qual(pkg, "ExactlyOneOf").Custom(multilineCall(), fields...),
	)
	return jen.Func().Id(datasourceType.String()+configValidatorsFuncSuffix).
		Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("client").Qual(avnGenPackage, "Client"),
		).
		Index().Qual(datasourcePkg, "ConfigValidator").
		Block(
			jen.Return(validators),
		)
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
func humanizeCodeList(args []string, union string) string {
	var list strings.Builder
	last := len(args) - 1
	for i, v := range args {
		switch i {
		case 0:
		case last:
			list.WriteString(union)
		default:
			list.WriteString(", ")
		}
		list.WriteString(fmt.Sprintf("`%s`", v))
	}
	return list.String()
}
