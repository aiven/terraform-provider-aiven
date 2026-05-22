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
	validateConfig             = "validateConfig"
	modifyPlan                 = "modifyPlan"
	resourceData               = "ResourceData"
	renameFieldsModifier       = "RenameFields"
	flattenModifier            = "flattenModifier"
	expandModifier             = "expandModifier"
	refreshStateWaiter         = "refreshStateWaiter"
	populateIDFieldsFromID     = "populateIDFieldsFromID"
	waitForDeletion            = "waitForDeletion"
)

// genViews generates CRUD views for the resource, skips disabled or undefined operations.
func genViews(def *Definition, item *Item) ([]jen.Code, error) {
	codes := make([]jen.Code, 0)
	for _, op := range listOperationTypes() {
		v, err := genGenericView(def, item, op)
		if err != nil {
			return nil, fmt.Errorf("%s view error: %w", op, err)
		}
		if v != nil {
			codes = append(codes, v)
		}
	}

	if def.Resource != nil && def.IDAttributeComposed.PopulateFieldsFromID {
		codes = append(codes, genPopulateIDFieldsFromID(def))
	}

	builders := make([]jen.Code, 0)
	if def.Resource != nil {
		builders = append(builders, genNewResource(resourceType, def, false))
	}

	if def.Datasource != nil {
		hasLookup := def.DatasourceLookupOp() != nil
		if hasLookup {
			codes = append(codes, datasourceLookupValidators(def, item))
		}

		builders = append(builders, genNewResource(datasourceType, def, hasLookup))
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
		if entity == datasourceType {
			if v.Type != OperationRead {
				// Datasource only supports Read operation
				continue
			}

			if v.DatasourceLookup {
				// Lookup ops are inlined into readView's id-empty branch; no separate
				// wiring is needed (and they're invalid on resources).
				continue
			}
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

		if def.Resource.IgnoreAlreadyExists {
			values["IgnoreAlreadyExists"] = jen.True()
		}

		if def.Resource.TerminationProtection {
			values["TerminationProtection"] = jen.True()
		}

		if def.Resource.ValidateConfig {
			values["ValidateConfig"] = jen.Id(validateConfig)
		}

		if def.Resource.ModifyPlan {
			values["ModifyPlan"] = jen.Id(modifyPlan)
		}
	}

	if hasConfigValidators {
		values["ConfigValidators"] = jen.Id(entity.String() + configValidatorsFuncSuffix)
	}

	title := entity.Title() + optionsSuffix
	returnValue := jen.Qual(adapterPackage, title).Values(dictFromMap(values, false))
	return jen.Var().Id(title).Op("=").Add(returnValue)
}

func genGenericView(def *Definition, item *Item, operation OperationType) (jen.Code, error) {
	operations := make([]*Operation, 0)
	for _, op := range def.Operations {
		if op.Type == operation && !op.DisableView && !op.DatasourceLookup {
			operations = append(operations, op)
		}
	}

	if len(operations) == 0 {
		return nil, nil
	}

	// readView absorbs the datasourceLookup op as its id-empty branch:
	// when the canonical id is missing, run the lookup (which either returns
	// the flattened match or sets id and falls through to the canonical read).
	var lookupOp *Operation
	if operation == OperationRead {
		lookupOp = def.DatasourceLookupOp()
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
		if operation == OperationRead {
			emitDatasourceLookupIDSplit(g, def)
		}

		emitPopulateIDFieldsFromID(g, def, operation)

		if operation == OperationRead {
			emitPlanModifier(g, def)
		}

		if lookupOp != nil {
			var lookupErr error
			g.If(
				jen.Id("d").Dot("Get").Call(jen.Lit(def.DatasourceCanonicalID())).Op(".").Parens(jen.String()).Op("==").Lit(""),
			).BlockFunc(func(g *jen.Group) {
				lookupErr = genGenericViewOperation(g, 0, 1, def, item, lookupOp)
			})
			if lookupErr != nil {
				errM = multierror.Append(errM, fmt.Errorf("%s view error: %w", lookupOp.ID, lookupErr))
			}
		}

		for i, op := range operations {
			err := genGenericViewOperation(g, i, len(operations), def, item, op)
			if err != nil {
				errM = multierror.Append(errM, fmt.Errorf("%s view error: %w", op.ID, err))
			}
		}
	})

	return view, errM.ErrorOrNil()
}

func emitDatasourceLookupIDSplit(g *jen.Group, def *Definition) {
	composed := def.DatasourceLookupIDComposed()
	if len(composed) == 0 {
		return
	}

	lookupID := def.DatasourceLookupID()
	format := strings.Join(composed, "/")
	g.If(
		jen.List(jen.Id("_"), jen.Id("exists")).Op(":=").Id("d").Dot("Schema").Call().Dot("Properties").Index(jen.Lit(lookupID)),
		jen.Id("exists"),
	).BlockFunc(func(g *jen.Group) {
		g.If(
			jen.List(jen.Id("v"), jen.Id("ok")).Op(":=").Id("d").Dot("GetOk").Call(jen.Lit(lookupID)),
			jen.Id("ok"),
		).BlockFunc(func(g *jen.Group) {
			g.List(jen.Id("parts"), jen.Err()).Op(":=").Qual(schemautilPackage, "SplitResourceID").
				Call(jen.Id("v").Op(".").Parens(jen.String()), jen.Lit(len(composed)))
			g.If(jen.Err().Op("!=").Nil()).Block(
				jen.Return().Qual("fmt", "Errorf").Call(
					jen.Lit(fmt.Sprintf("invalid %s, expected format %s: %%w", lookupID, format)),
					jen.Err(),
				),
			)

			for i, name := range composed {
				g.If(
					jen.Err().Op(":=").Id("d").Dot("Set").Call(jen.Lit(name), jen.Id("parts").Index(jen.Lit(i))),
					jen.Err().Op("!=").Nil(),
				).Block(jen.Return().Err())
			}
		})
	})
}

// emitPlanModifier prepends the planModifier call to a Read-style view body.
// No-op when the definition does not opt in via planModifier: true.
func emitPlanModifier(g *jen.Group, def *Definition) {
	if !def.PlanModifier {
		return
	}
	g.Err().Op(":=").Id(planModifier).Call(
		jen.Id("ctx"),
		jen.Id("client"),
		jen.Id("d"),
	)
	g.If(jen.Err().Op("!=").Nil()).Block(jen.Return().Err())
}

func shouldPopulateIDFieldsFromID(def *Definition, operation OperationType) bool {
	return def.Resource != nil &&
		def.IDAttributeComposed.PopulateFieldsFromID &&
		(operation == OperationRead || operation == OperationDelete)
}

func emitPopulateIDFieldsFromID(g *jen.Group, def *Definition, operation OperationType) {
	if !shouldPopulateIDFieldsFromID(def, operation) {
		return
	}
	g.If(
		jen.Err().Op(":=").Id(populateIDFieldsFromID).Call(jen.Id("d")),
		jen.Err().Op("!=").Nil(),
	).Block(jen.Return().Err())
}

func genPopulateIDFieldsFromID(def *Definition) jen.Code {
	return jen.Func().Id(populateIDFieldsFromID).
		Params(jen.Id("d").Qual(adapterPackage, resourceData)).
		Error().
		Block(
			jen.If(jen.Id("d").Dot("ID").Call().Op("==").Lit("")).Block(jen.Return().Nil()),
			jen.Id("fields").Op(":=").Id(funcIDFields).Call(),
			jen.List(jen.Id("chunks"), jen.Err()).Op(":=").Qual(schemautilPackage, "SplitResourceID").Call(
				jen.Id("d").Dot("ID").Call(),
				jen.Len(jen.Id("fields")),
			),
			jen.If(jen.Err().Op("!=").Nil()).Block(
				jen.Return().Qual("fmt", "Errorf").Call(
					jen.Lit("invalid "+def.typeName+" id %q: %w"),
					jen.Id("d").Dot("ID").Call(),
					jen.Err(),
				),
			),
			jen.For(jen.List(jen.Id("i"), jen.Id("field")).Op(":=").Range().Id("fields")).Block(
				jen.If(jen.Id("d").Dot("Get").Call(jen.Id("field")).Op(".").Parens(jen.String()).Op("!=").Lit("")).Block(jen.Continue()),
				jen.If(
					jen.Err().Op(":=").Id("d").Dot("Set").Call(jen.Id("field"), jen.Id("chunks").Index(jen.Id("i"))),
					jen.Err().Op("!=").Nil(),
				).Block(jen.Return().Err()),
			),
			jen.Return().Nil(),
		)
}

func genWaitForDeletionCall() *jen.Statement {
	return jen.Id(waitForDeletion).Call(jen.Id("ctx"), jen.Id("client"), jen.Id("d"))
}

func genReturnErrOrWaitForDeletion(g *jen.Group, operation *Operation) {
	if !operation.WaitForDeletion {
		g.Return().Err()
		return
	}

	g.If(jen.Err().Op("!=").Nil()).Block(jen.Return().Err())
	g.Return().Add(genWaitForDeletionCall())
}

func genGenericViewOperation(g *jen.Group, funcIndex, funcCount int, def *Definition, item *Item, operation *Operation) error {
	if operation.DatasourceLookup && len(operation.ResultListLookupKeys) == 0 {
		return fmt.Errorf("datasourceLookup operation %q must declare resultListLookupKeys for filtering", operation.ID)
	}

	inPath := def.Operations.AppearsInID(operation.ID, operation.Type, PathParameter)
	inRequest := def.Operations.AppearsInID(operation.ID, operation.Type, RequestBody)
	inResponse := def.Operations.AppearsInID(operation.ID, operation.Type, ResponseBody)

	// Properties must be sorted for consistent output
	properties := slices.Collect(maps.Values(item.Properties))
	sort.Slice(properties, func(i, j int) bool {
		return properties[i].Name < properties[j].Name
	})

	hasRequest := len(filterAppearsIn(properties, inRequest)) > 0
	// Ops with resultIDField never merge their response into root.Properties, but the
	// emitted view still consumes the response (FindOne, then d.Set id + readView).
	hasResponse := len(filterAppearsIn(properties, inResponse)) > 0 || (operation.ResultIDField != "" && operation.Response != nil)

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
			if code := renameFieldsArg(def.Rename, operation.Request, true); code != nil {
				g.Add(code)
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
	case mustReturnResult && !operation.WaitForDeletion:
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
		if operation.WaitForDeletion {
			g.If(jen.Err().Op("!=").Nil()).Block(jen.Return().Err())
			g.Return().Add(genWaitForDeletionCall())
		}
		return nil
	}

	if !hasResponse || operation.Type == OperationDelete {
		genReturnErrOrWaitForDeletion(g, operation)
		return nil
	}

	// Reads the response from the previous call
	g.If(jen.Err().Op("!=").Nil()).Block(jen.Return().Err())
	flatten := func(rsp jen.Code) jen.Code {
		v := jen.Id("d").Dot("Flatten").CallFunc(func(g *jen.Group) {
			g.Add(rsp)

			if code := renameFieldsArg(def.Rename, operation.Response, false); code != nil {
				g.Add(code)
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

	// rspCode is the response expression, optionally traversed by
	// ResultKeyField (e.g. rsp.Aws or rsp.ConnectionPools) when the Go
	// client doesn't strip an extra wrapper exposed by the API.
	rspCode := jen.Id(rspName)
	if operation.ResultKeyField != "" {
		rspCode.Dot(operation.ResultKeyField)
	}

	// Wraps the result into a map by the specified key
	if operation.ResultToKey != "" {
		m := jen.Op("&").Map(jen.String()).Any()
		g.Add(flatten(m.Values(jen.Dict{jen.Lit(operation.ResultToKey): rspCode})))
		return nil
	}

	// Flattens the response directly, because it is not a list but an object
	if len(operation.ResultListLookupKeys) == 0 {
		g.Add(flatten(rspCode))
		return nil
	}

	// Finds a match in the list.
	// E.g. rsp[0].foo == state.Foo && rsp[0].bar == state.Bar ...

	var fieldsMatch jen.Statement
	for i, key := range sortedKeys(operation.ResultListLookupKeys) {
		fieldName := operation.ResultListLookupKeys[key]
		if i > 0 {
			fieldsMatch.Op("&&")
		}

		field, ok := item.Properties[fieldName]
		if !ok {
			return fmt.Errorf("unknown lookup key %q in result list for operation %q", fieldName, operation.ID)
		}

		fieldsMatch.Qual(adapterPackage, "Equal").Call(
			rspCode.Clone().Index(jen.Id("i")).Dot(key),
			jen.Id("d").Dot("Get").Call(jen.Lit(field.Name)),
		)
	}

	const matchName = "match"
	g.List(jen.Id(matchName), jen.Err()).Op(":=").
		Qual(adapterPackage, "FindOne").
		CallFunc(func(g *jen.Group) {
			g.Add(rspCode)
			g.Func().Params(jen.Id("i").Int()).Bool().Block(jen.Return(&fieldsMatch))
		})
	g.If(jen.Err().Op("!=").Nil()).Block(
		jen.Return().Qual("fmt", "Errorf").Call(
			jen.Lit(fmt.Sprintf("lookup `%s` by %s: %%w", def.typeName, lookupFieldsList(operation))),
			jen.Err(),
		),
	)

	// resultIDField: the lookup endpoint only resolves the primary id; control then
	// falls through to the canonical read body (emitted right after the lookup block
	// in readView). Used when the lookup response shape differs from the read response.
	if operation.DatasourceLookup && operation.ResultIDField != "" {
		g.If(
			jen.Err().Op(":=").Id("d").Dot("Set").Call(
				jen.Lit(def.DatasourceCanonicalID()),
				jen.Id(matchName).Dot(operation.ResultIDField),
			),
			jen.Err().Op("!=").Nil(),
		).Block(jen.Return().Err())
		return nil
	}

	g.Add(flatten(jen.Op("&").Id(matchName)))

	return nil
}

// lookupFieldsList renders an op's ResultListLookupKeys tf-attribute values as a
// backticked human-readable list joined by " and " (or " or " when OR mode is set).
func lookupFieldsList(op *Operation) string {
	values := lo.Values(op.ResultListLookupKeys)
	slices.Sort(values)
	return humanizeCodeList(values, " and ")
}

// datasourceLookupValidators emits the config-level validation for a datasource lookup.
// With the default leaf id lookup it preserves the existing id-or-alt-key contract.
// With a composed public lookup alias it also requires the path parameters needed
// by the alternative lookup branch and conflicts those parameters with the alias.
func datasourceLookupValidators(def *Definition, item *Item) jen.Code {
	id := def.DatasourceLookupID()
	composedOf := def.DatasourceLookupComposedOf()
	pkg := datasourceType.Import(entityValidatorPkgFmt)

	matchRoot := func(name string) jen.Code {
		return jen.Qual(pathPackage, "MatchRoot").Call(jen.Lit(name))
	}

	validators := make([]jen.Code, 0, 2)
	validators = append(validators, jen.Qual(pkg, "ExactlyOneOf").Custom(
		multilineCall(),
		matchRoot(id),
		matchRoot(composedOf[0]),
	))

	togetherFields := composedOf
	if def.hasDatasourceLookupAlias() {
		op := def.DatasourceLookupOp()
		lookupPathParams := filterAppearsIn(
			slices.Collect(maps.Values(item.Properties)),
			def.Operations.AppearsInID(op.ID, op.Type, PathParameter),
		)
		for _, v := range lookupPathParams {
			togetherFields = append(togetherFields, v.Name)
		}
		slices.Sort(togetherFields)
		togetherFields = slices.Compact(togetherFields)
	}

	if len(togetherFields) > 1 {
		together := make([]jen.Code, len(togetherFields))
		for i, v := range togetherFields {
			together[i] = matchRoot(v)
		}
		validators = append(validators, jen.Qual(pkg, "RequiredTogether").Custom(multilineCall(), together...))
	}

	if def.hasDatasourceLookupAlias() {
		conflictsWithAlias := make([]jen.Code, 0)
		for _, v := range def.DatasourceLookupIDComposed() {
			if v == def.DatasourceCanonicalID() || v == id {
				continue
			}
			prop, ok := item.Properties[v]
			if !ok || !prop.IsOptional(def, datasourceType) {
				continue
			}
			conflictsWithAlias = append(conflictsWithAlias, matchRoot(v))
		}
		if len(conflictsWithAlias) > 0 {
			validators = append(validators, jen.Qual(pkg, "Conflicting").Custom(
				multilineCall(),
				append([]jen.Code{matchRoot(id)}, conflictsWithAlias...)...,
			))
		}
	}

	list := jen.Index().Qual(datasourcePkg, "ConfigValidator").Custom(multilineValues(), validators...)
	return jen.Func().Id(datasourceType.String()+configValidatorsFuncSuffix).
		Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("client").Qual(avnGenPackage, "Client"),
		).
		Index().Qual(datasourcePkg, "ConfigValidator").
		Block(
			jen.Return(list),
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

// renameFieldsArg builds the adapter.RenameFields(...) modifier for the subset of
// def.Rename that the schema actually carries. Returns nil when nothing applies.
// forExpand=true emits {tfName: jsonName} (Expand: TF -> request); otherwise
// {jsonName: tfName} (Flatten: response -> TF).
func renameFieldsArg(rename map[string]string, schema *OASchema, forExpand bool) jen.Code {
	if len(rename) == 0 || schema == nil {
		return nil
	}
	dict := make(jen.Dict)
	for jsonName, tfName := range rename {
		if _, ok := schema.Properties[jsonName]; !ok {
			continue
		}
		if forExpand {
			dict[jen.Lit(tfName)] = jen.Lit(jsonName)
		} else {
			dict[jen.Lit(jsonName)] = jen.Lit(tfName)
		}
	}
	if len(dict) == 0 {
		return nil
	}
	return jen.Qual(adapterPackage, renameFieldsModifier).Call(
		jen.Map(jen.String()).String().Values(dict),
	)
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
