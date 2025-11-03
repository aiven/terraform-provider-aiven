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
	codes := make([]jen.Code, 0)
	for _, op := range listOperations() {
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
		if entity == datasourceType && v != OperationRead {
			// Datasource only supports Read operation
			continue
		}
		values[firstUpper(v)] = jen.Id(fmt.Sprintf("%s%s", v, viewSuffix))
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

func genGenericView(item *Item, def *Definition, operation Operation) (jen.Code, error) {
	funcBlocks := make([]jen.Code, 0)
	count := 0
	for opID, op := range def.Operations {
		if op == operation {
			if count > 0 {
				funcBlocks = append(
					funcBlocks,
					jen.If(jen.Id("diags").Dot("HasError").Call()).Block(jen.Return().Id("diags")),
				)
			}
			thisBlocks, err := genGenericViewOperation(item, def, operation, opID)
			if err != nil {
				return nil, fmt.Errorf("generating operation ID %q view: %w", opID, err)
			}
			funcBlocks = append(funcBlocks, jen.Func().Params().Block(thisBlocks...).Call())
			count++
		}
	}

	if count == 0 {
		return nil, nil
	}

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

	// The return block
	var blocks []jen.Code
	blocks = append(blocks, jen.Var().Id("diags").Qual(diagPackage, "Diagnostics"))
	blocks = append(blocks, funcBlocks...)
	blocks = append(blocks, jen.Return().Id("diags"))

	return jen.Func().Id(fmt.Sprintf("%s%s", operation, viewSuffix)).
		Params(opFuncParams...).
		Qual(diagPackage, "Diagnostics").
		Block(blocks...), nil
}

func genGenericViewOperation(item *Item, def *Definition, operation Operation, operationID string) ([]jen.Code, error) {
	inPath := def.Operations.AppearsInID(operationID, PathParameter)
	inRequest := def.Operations.AppearsInID(operationID, RequestBody)
	inResponse := def.Operations.AppearsInID(operationID, ResponseBody)

	// Properties must be sorted for consistent output
	properties := slices.Collect(maps.Values(item.Properties))
	sort.Slice(properties, func(i, j int) bool {
		return properties[i].Name < properties[j].Name
	})

	blocks := make([]jen.Code, 0)
	hasRequest := len(filterAppearsIn(properties, inRequest)) > 0
	hasResponse := len(filterAppearsIn(properties, inResponse)) > 0

	// Generates the code that converts the TF model into the Client request struct
	if hasRequest {
		state := jen.Id("state")
		if operation == OperationCreate {
			state = jen.Nil()
		}

		pkgName := avnGenHandlerPackage + def.ClientHandler
		blocks = append(
			blocks,
			// Creates the request var
			jen.Var().Id("req").Qual(pkgName, operationID+"In"),

			// Expands data into the request
			jen.Id("diags").Dot("Append").Call(
				jen.Id(expandDataFunc).
					Call(
						jen.Id("ctx"),
						jen.Id("plan"),
						state,
						jen.Id("&req"),
					).
					Op("..."),
			),

			// If HasError, then return
			jen.If(jen.Id("diags").Dot("HasError").Call()).Block(jen.Return()),
		)
	}

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

		idFields := filterAppearsIn(properties, inPath)
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
				jen.Id("diags").Dot("Append").Call(
					jen.Qual(errMsgPackage, "FromError").Call(
						jen.Lit(fmt.Sprintf(`%s Error`, firstUpper(operationID))),
						jen.Err(),
					),
				),
				jen.Return(),
			)
	}())

	// Reads the response from the previous call
	if hasResponse && operation != OperationDelete {
		state := "plan"
		if operation == OperationRead {
			state = "state"
		}
		blocks = append(
			blocks,
			jen.Id("diags").Dot("Append").Call(
				jen.Id(flattenDataFunc).
					Call(
						jen.Id("ctx"),
						jen.Id(state),
						jen.Id("rsp"),
					).
					Op("..."),
			),
		)
	}

	return blocks, nil
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
