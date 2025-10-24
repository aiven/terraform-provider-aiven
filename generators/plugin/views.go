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

		v, err := genGenericView(item, def, op)
		if err != nil {
			return nil, fmt.Errorf("%s view error: %w", op, err)
		}
		if v != nil {
			codes = append(codes, v)
		}
	}

	if len(codes) > 0 && def.DisableViews == nil {
		builders := make([]jen.Code, 0)
		if def.Resource != nil {
			builders = append(builders, genNewResource(resourceType, def))
		}

		if def.Datasource != nil {
			builders = append(builders, genNewResource(datasourceType, def))
		}

		codes = append(builders, codes...)
	}

	return codes, nil
}

// genNewResource generates NewResource or NewDatasource function
func genNewResource(entity entityType, def *Definition) jen.Code {
	entityTitle := firstUpper(entity)
	values := jen.Dict{
		jen.Id("TypeName"): jen.Id(aivenNameConst),
		jen.Id("Schema"):   jen.Id(fmt.Sprintf("new%sSchema", entityTitle)),
	}

	for _, v := range def.Operations {
		if entity == datasourceType && v != OperationRead {
			// Datasource only supports Read operation
			continue
		}
		values[jen.Id(firstUpper(v))] = jen.Id(fmt.Sprintf("%s%s", v, viewSuffix))
	}

	entityName := string(entity)
	entityInterface := "DataSource" // It is called "DataSource" not "Datasource"
	entityPackage := datasourcePackage
	if entity == resourceType {
		entityInterface = entityTitle
		entityPackage = resourcePackage
		values[jen.Id("IDFields")] = jen.Id(funcIDFields).Call()
	}

	funcName := "New" + entityTitle
	genericType := jen.List(jen.Op("*").Id(entityName+"Model"), jen.Id(tfRootModel))
	returnValue := jen.Qual(adapterPackage, entityTitle+"Options").Index(genericType).Values(values)
	return jen.Func().Id(funcName).Params().
		Qual(entityPackage, entityInterface).
		Block(jen.Return().Qual(adapterPackage, funcName).Call(returnValue))
}

func genGenericView(item *Item, def *Definition, operation Operation) (jen.Code, error) {
	inPath := PathParameter
	var inView, inRequest, inResponse AppearsIn
	switch operation {
	case OperationDelete:
		inView = DeleteHandler
		inResponse = DeleteResponseBody
	case OperationRead:
		inView = ReadHandler
		inResponse = ReadResponseBody
	case OperationUpdate:
		inView = UpdateHandler
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

	var operationID string
	for opID, op := range def.Operations {
		if op == operation {
			operationID = opID
		}
	}
	if operationID == "" {
		// Create, Read, Update, or Delete operation is not defined for the item
		return nil, nil
	}

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
				jen.Id("diags").Dot("AddError").Call(
					jen.Lit(fmt.Sprintf(`%s Error`, firstUpper(operationID))),
					jen.Err().Dot("Error").Call(),
				),
				jen.Return().Id("diags"),
			).Line()
	}())

	// Reads the response from the previous call
	if hasResponse {
		switch operation {
		case OperationCreate, OperationUpdate:
			blocks = append(
				blocks,
				jen.Comment("The response may contain values that don't exist in the Read operation."),
				jen.Comment("Additionally, the Read operation needs ID fields to format the URL, which may come from this response."),
				jen.Id("diags").Dot("Append").Call(
					jen.Id(flattenDataFunc).Call(
						jen.Id("ctx"),
						jen.Id("plan"),
						jen.Id("rsp"),
					).Op("..."),
				),
				jen.If(jen.Id("diags").Dot("HasError").Call()).
					Block(jen.Return().Id("diags")).Line(),
			)
		}
	}

	// Additionally, calls the Read view after Create or Update
	// to ensure the state is in sync with the server
	blocks = append(blocks, func() jen.Code {
		switch operation {
		case OperationRead:
			return jen.
				Return().Append(
				jen.Id("diags"),
				jen.Id(flattenDataFunc).Call(
					jen.Id("ctx"),
					jen.Id("state"),
					jen.Id("rsp"),
				).Op("..."),
			)
		case OperationCreate, OperationUpdate:
			// If the read view was disabled, we can't call it.
			if !slices.Contains(def.DisableViews, OperationRead) {
				return jen.
					Comment("Reads the remote state to sync and detect drift.").Line().
					Return().Append(
					jen.Id("diags"),
					jen.Id(string(OperationRead)+viewSuffix).Call(
						jen.Id("ctx"),
						jen.Id("client"),
						jen.Id("plan"),
					).Op("..."),
				)
			}
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
