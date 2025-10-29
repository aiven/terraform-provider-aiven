package main

import (
	"fmt"
	"maps"
	"slices"
	"sort"

	"github.com/dave/jennifer/jen"
)

const (
	avnGenHandlerPackage = "github.com/aiven/go-client-codegen/handler/"
	viewSuffix           = "View"
	optionsSuffix        = "Options"
)

// genViews generates CRUD views for the resource, skips disabled or undefined operations.
func genViews(item *Item, def *Definition) ([]jen.Code, error) {
	operations := []Operation{
		OperationCreate,
		OperationUpdate,
		OperationRead,
		OperationDelete,
	}

	codes := make([]jen.Code, 0)
	for _, op := range operations {
		if slices.Contains(def.DisableViews, op) {
			continue
		}

		for opID, o := range def.Operations {
			if o != op {
				continue
			}

			v, err := genGenericView(item, def, op, opID)
			if err != nil {
				return nil, fmt.Errorf("%s view error: %w", op, err)
			}
			if v != nil {
				codes = append(codes, v)
			}
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
	values := jen.Dict{
		jen.Id("TypeName"): jen.Id(typeName),
		jen.Id("Schema"):   jen.Id(string(entity) + schemaSuffix),
	}

	for _, v := range def.Operations {
		if entity == datasourceType && v != OperationRead {
			// Datasource only supports Read operation
			continue
		}
		values[jen.Id(firstUpper(v))] = jen.Id(fmt.Sprintf("%s%s", v, viewSuffix))
	}

	entityName := string(entity)
	if entity == resourceType {
		values[jen.Id("IDFields")] = jen.Id(funcIDFields).Call()
		values[jen.Id("RefreshState")] = jen.Lit(def.RefreshState)
	}

	title := entity.Title() + optionsSuffix
	genericType := jen.List(jen.Op("*").Id(entityName+"Model"), jen.Id(tfRootModel))
	returnValue := jen.Qual(adapterPackage, title).Index(genericType).Values(values)
	return jen.Var().Id(title).Op("=").Add(returnValue)
}

func genGenericView(item *Item, def *Definition, operation Operation, operationID string) (jen.Code, error) {
	var inView, inPath, inRequest, inResponse AppearsIn
	switch operation {
	case OperationDelete:
		inView = DeleteHandler
		inPath = DeletePathParameter
		inResponse = DeleteResponseBody
	case OperationRead:
		inView = ReadHandler
		inPath = ReadPathParameter
		inResponse = ReadResponseBody
	case OperationUpdate:
		inView = UpdateHandler
		inPath = UpdatePathParameter
		inRequest = UpdateRequestBody
		inResponse = UpdateResponseBody
	case OperationCreate:
		inView = CreateHandler
		inPath = CreatePathParameter
		inRequest = CreateRequestBody
		inResponse = CreateResponseBody
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}

	// Properties must be sorted for consistent output
	properties := slices.Collect(maps.Values(item.Properties))
	sort.Slice(properties, func(i, j int) bool {
		return properties[i].Name < properties[j].Name
	})

	blocks := make([]jen.Code, 0)
	hasRequest := len(filterAppearsIn(properties, inRequest)) > 0
	hasResponse := len(filterAppearsIn(properties, inResponse)) > 0

	// Generates the code that converts the TF model into the Client request struct
	blocks = append(blocks, func() jen.Code {
		req := jen.Var().Id("diags").Qual(diagPackage, "Diagnostics")
		if !hasRequest {
			return req
		}

		state := jen.Id("state")
		if operation == OperationCreate {
			state = jen.Nil()
		}

		pkgName := avnGenHandlerPackage + def.ClientHandler
		req = jen.Var().Id("req").Qual(pkgName, operationID+"In")
		req.Line().
			Id("diags").Op(":=").Id(expandDataFunc).
			Call(
				jen.Id("ctx"),
				jen.Id("plan"),
				state,
				jen.Id("&req"),
			).Line().
			If(jen.Id("diags").Dot("HasError").Call()).
			Block(jen.Return().Id("diags")).Line()
		return req
	}())

	// Generates the Client request block
	blocks = append(blocks, func() jen.Code {
		// Gathers parameters: ctx and path parameters to call the client method
		clientParams := []jen.Code{
			jen.Id("ctx"),
		}

		state := "state"
		if operation == OperationCreate {
			state = "plan"
		}

		idFields := filterAppearsIn(properties, inView, inPath)
		sort.Slice(idFields, func(i, j int) bool {
			return idFields[i].IDAttributePosition < idFields[j].IDAttributePosition
		})

		for _, v := range idFields {
			clientParams = append(clientParams, jen.Id(state).Dot(v.GoFieldName()).Dot("Value"+v.TFType()).Call())
		}

		if hasRequest {
			clientParams = append(clientParams, jen.Id("&req"))
		}

		clientResult := jen.Err()
		if hasResponse {
			// If there is a response for Delete operation, then renders "_, err :="
			// otherwise "rsp, err :="
			if operation == OperationDelete {
				clientResult = jen.List(jen.Id("_"), jen.Err())
			} else {
				clientResult = jen.List(jen.Id("rsp"), jen.Err())
			}
		}

		return clientResult.Op(":=").
			Id("client").Dot(operationID).Call(clientParams...).Line().
			If(jen.Err().Op("!=").Nil()).
			Block(
				jen.Return().Append(
					jen.Id("diags"),
					jen.Qual(errMsgPackage, "FromError").Call(
						jen.Lit(fmt.Sprintf(`%s Error`, firstUpper(operationID))),
						jen.Err(),
					),
				),
			).Line()
	}())

	// Reads the response from the previous call
	blocks = append(blocks, func() jen.Code {
		switch operation {
		case OperationRead, OperationCreate, OperationUpdate:
			if !hasResponse {
				break
			}

			state := "plan"
			if operation == OperationRead {
				state = "state"
			}

			return jen.Return().Append(
				jen.Id("diags"),
				jen.Id(flattenDataFunc).Call(
					jen.Id("ctx"),
					jen.Id(state),
					jen.Id("rsp"),
				).Op("..."),
			)
		}

		return jen.Return().Id("diags")
	}())

	opFuncParams := []jen.Code{
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("client").Qual(avnGenPackage, "Client"),
	}

	// Generates the view function.
	// See ResourceOptions for parameter details.
	switch operation {
	case OperationCreate:
		opFuncParams = append(opFuncParams, jen.Id("plan").Op("*").Id(tfRootModel))
	case OperationUpdate:
		opFuncParams = append(
			opFuncParams,
			jen.List(jen.Id("plan"), jen.Id("state"), jen.Id("config")).Op("*").Id(tfRootModel),
		)
	default:
		opFuncParams = append(opFuncParams, jen.Id("state").Op("*").Id(tfRootModel))
	}

	return jen.Func().Id(fmt.Sprintf("%s%s", operation, viewSuffix)).
		Params(opFuncParams...).
		Qual(diagPackage, "Diagnostics").
		Block(blocks...), nil
}

// filterAppearsIn filters items by their appearance in Create/Update request, response, etc.
func filterAppearsIn(items []*Item, appearsIn ...AppearsIn) []*Item {
	var props []*Item
	for _, v := range items {
		if v.appearsIn(appearsIn...) {
			props = append(props, v)
		}
	}
	return props
}
