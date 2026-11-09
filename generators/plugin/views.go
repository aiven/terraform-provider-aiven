package main

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

const (
	viewSuffix                 = "View"
	optionsSuffix              = "Options"
	configValidatorsFuncSuffix = "ConfigValidators"
	readModifier               = "readModifier"
	validateConfig             = "validateConfig"
	modifyPlan                 = "modifyPlan"
	resourceData               = "ResourceData"
	renameFieldsModifier       = "RenameFields"
	flattenModifier            = "flattenModifier"
	expandModifier             = "expandModifier"
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

	paged, err := pagedOperation(def)
	if err != nil {
		return nil, err
	}
	if paged != nil {
		codes = append(codes, genPageView(def, item, paged))
		if len(paged.ResultListLookupKeys) > 0 {
			v, err := genPageMatch(def, item, paged)
			if err != nil {
				return nil, fmt.Errorf("%s page match error: %w", paged.ID, err)
			}
			codes = append(codes, v)
		}
	}

	builders := make([]jen.Code, 0)
	if def.Resource != nil {
		v, err := genNewResource(resourceType, def, item, false)
		if err != nil {
			return nil, fmt.Errorf("resource builder error: %w", err)
		}
		builders = append(builders, v)
	}

	if def.Datasource != nil {
		hasLookup := def.DatasourceLookupOp() != nil
		hasExactlyOneOf := len(def.Datasource.ExactlyOneOf) > 0
		hasConfigValidators := hasLookup || hasExactlyOneOf
		if hasConfigValidators {
			codes = append(codes, datasourceLookupValidators(def))
		}

		v, err := genNewResource(datasourceType, def, item, hasConfigValidators)
		if err != nil {
			return nil, fmt.Errorf("datasource builder error: %w", err)
		}
		builders = append(builders, v)
	}

	return append(builders, codes...), nil
}

// genNewResource generates NewResource or NewDatasource function
func genNewResource(entity entityType, def *Definition, item *Item, hasConfigValidators bool) (jen.Code, error) {
	// Resource might have multiple read operations.
	// We use here a dict, so we don't need to handle duplicates.
	values := map[string]jen.Code{
		"TypeName":       jen.Id(typeName),
		"Schema":         jen.Id(string(entity) + schemaSuffix),
		"SchemaInternal": jen.Id(fmt.Sprintf(schemaInternalFmt, entity)).Call(),
		"IdentityFields": stringSliceLiteral(identityFieldNames(item)),
	}

	if lo.FromPtr(def.Beta) {
		values["Beta"] = jen.True()
	}

	for _, v := range def.Operations {
		if entity.IsDataSource() {
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

	for _, setting := range []struct {
		name   string
		fields []string
	}{
		{"identityFields", def.IdentityFields},
		{"identityFieldsOptional", def.IdentityFieldsOptional},
	} {
		if setting.fields == nil {
			continue
		}
		if len(setting.fields) == 0 {
			return nil, fmt.Errorf("%s must not be empty; omit it when unused", setting.name)
		}
		identityFields := make(map[string]bool)
		for _, name := range setting.fields {
			if identityFields[name] {
				return nil, fmt.Errorf("%s contains duplicate field %q", setting.name, name)
			}
			identityFields[name] = true
			if setting.name == "identityFieldsOptional" {
				// Trailing, so a shorter string ID unambiguously omits the last fields.
				if i := slices.Index(def.IdentityFields, name); i < 0 || i < len(def.IdentityFields)-len(setting.fields) {
					return nil, fmt.Errorf("%s requires %q to be listed in identityFields after the required fields", setting.name, name)
				}
				continue
			}
			prop, ok := item.Properties[name]
			if !ok {
				return nil, fmt.Errorf("%s requires %q attribute in schema", setting.name, name)
			}
			if prop.Type != SchemaTypeString {
				return nil, fmt.Errorf("%s requires %q to be a string attribute, got %q", setting.name, name, prop.Type)
			}
			if prop.WriteOnly {
				return nil, fmt.Errorf("%s cannot use write-only attribute %q", setting.name, name)
			}
		}
	}
	if len(def.IdentityFieldsOptional) > 0 {
		values["IdentityFieldsOptional"] = stringSliceLiteral(def.IdentityFieldsOptional)
	}

	if entity.isResource() {
		res := def.Resource
		// An updatable ID part changes identity in place; createRootItem drops UseStateForUnknown on id for the same reason.
		if identityFieldsMutable(def, item) {
			values["MutableIdentity"] = jen.True()
		}

		refresh, del := lo.FromPtr(res.RefreshState), lo.FromPtr(res.DeleteState)
		if refresh.Desired != nil && len(refresh.Desired) == 0 {
			return nil, errors.New("refreshState.desired must not be empty; use refreshState: true")
		}
		if refresh.Failed != nil && len(refresh.Failed) == 0 {
			return nil, errors.New("refreshState.failed must not be empty")
		}
		if del.Desired != nil && len(del.Desired) == 0 {
			return nil, errors.New("deleteState.desired must not be empty; use deleteState: true")
		}
		if len(refresh.Failed) > 0 && len(refresh.Desired) == 0 {
			return nil, errors.New("refreshState.failed requires refreshState.desired")
		}

		switch res.AlreadyExists {
		case "", AlreadyExistsAdopt, AlreadyExistsCheck:
		default:
			return nil, fmt.Errorf("alreadyExists must be %q or %q, got %q", AlreadyExistsAdopt, AlreadyExistsCheck, res.AlreadyExists)
		}
		if res.AlreadyExists != "" && res.RefreshState == nil {
			return nil, errors.New("alreadyExists requires refreshState")
		}

		needsStateAttribute := len(refresh.Desired) > 0 || len(del.Desired) > 0
		if res.StateAttribute != "" && !needsStateAttribute {
			return nil, errors.New("stateAttribute requires refreshState.desired or deleteState.desired")
		}
		if needsStateAttribute && res.StateAttribute == "" {
			return nil, errors.New("stateAttribute is required")
		}

		var stateProp *Item
		if res.StateAttribute != "" {
			var ok bool
			stateProp, ok = item.Properties[res.StateAttribute]
			if !ok {
				return nil, fmt.Errorf("stateAttribute %q is not present in schema", res.StateAttribute)
			}
		}

		if res.RefreshState != nil {
			fields := jen.Dict{}
			if len(refresh.Desired) > 0 {
				if !stateProp.IsComputed(def, entity) || !stateProp.IsReadOnly(def, entity) {
					return nil, fmt.Errorf("stateAttribute %q must be computed-only", res.StateAttribute)
				}
				if err := validateStateAttributeValues(res.StateAttribute, "refreshState.desired", refresh.Desired, stateProp); err != nil {
					return nil, err
				}
				if err := validateStateAttributeValues(res.StateAttribute, "refreshState.failed", refresh.Failed, stateProp); err != nil {
					return nil, err
				}
				for _, failed := range refresh.Failed {
					if slices.Contains(refresh.Desired, failed) {
						return nil, fmt.Errorf(
							"stateAttribute %q value %q cannot be both refreshState.desired and refreshState.failed",
							res.StateAttribute,
							failed,
						)
					}
				}

				fields = jen.Dict{
					jen.Id("Attribute"): jen.Lit(res.StateAttribute),
					jen.Id("Desired"):   stringSliceLiteral(refresh.Desired),
				}
				if len(refresh.Failed) > 0 {
					fields[jen.Id("Failed")] = stringSliceLiteral(refresh.Failed)
				}
			}

			values["RefreshState"] = jen.Op("&").
				Qual(adapterPackage, "RefreshStateCondition").
				Values(fields)

			if refresh.Delay != 0 {
				values["RefreshStateDelay"] = jen.Qual(adapterPackage, "MustParseDuration").Call(jen.Lit(refresh.Delay.String()))
			}
		}

		if res.DeleteState != nil {
			fields := jen.Dict{}
			if len(del.Desired) > 0 {
				if err := validateStateAttributeValues(res.StateAttribute, "deleteState.desired", del.Desired, stateProp); err != nil {
					return nil, err
				}
				fields[jen.Id("Attribute")] = jen.Lit(res.StateAttribute)
				fields[jen.Id("Desired")] = stringSliceLiteral(del.Desired)
			}
			values["DeleteState"] = jen.Op("&").Qual(adapterPackage, "DeleteStateOptions").Values(fields)
		}

		if def.Resource.RemoveMissing {
			values["RemoveMissing"] = jen.True()
		}

		if res.AlreadyExists != "" {
			values["IgnoreAlreadyExists"] = jen.True()
		}

		if res.AlreadyExists == AlreadyExistsCheck {
			v, err := genConflictOptions(def, item)
			if err != nil {
				return nil, fmt.Errorf("alreadyExists: check: %w", err)
			}
			values["Conflict"] = v
		}

		if def.Resource.TerminationProtection {
			values["TerminationProtection"] = jen.True()
		}

		switch {
		case def.Resource.ValidateConfig && def.UserConfigDiscriminator != "":
			return nil, fmt.Errorf("validateConfig: call adapter.%s from it, userConfigDiscriminator can't set it", validateUserConfig)
		case def.Resource.ValidateConfig:
			values["ValidateConfig"] = jen.Id(validateConfig)
		case def.UserConfigDiscriminator != "":
			values["ValidateConfig"] = jen.Qual(adapterPackage, validateUserConfig).Call(jen.Lit(def.UserConfigDiscriminator))
		}

		kinds := requiredUserConfigTypes(item)
		switch {
		case def.Resource.ModifyPlan && len(kinds) > 0:
			return nil, fmt.Errorf("modifyPlan: userConfigDiscriminator sets it to adapter.%s for %s", requireUserConfig, strings.Join(kinds, ", "))
		case def.Resource.ModifyPlan:
			values["ModifyPlan"] = jen.Id(modifyPlan)
		case len(kinds) > 0:
			values["ModifyPlan"] = jen.Qual(adapterPackage, requireUserConfig).CallFunc(func(g *jen.Group) {
				g.Lit(def.UserConfigDiscriminator)
				for _, k := range kinds {
					g.Lit(k)
				}
			})
		}
	}

	if len(def.RequireServiceReady) > 0 {
		for _, name := range []string{"project", "service_name"} {
			prop, ok := item.Properties[name]
			if !ok {
				return nil, fmt.Errorf("requireServiceReady requires %q attribute in schema", name)
			}
			if prop.Type != SchemaTypeString {
				return nil, fmt.Errorf("requireServiceReady requires %q to be a string attribute, got %q", name, prop.Type)
			}
		}
		values["RequireServiceReady"] = jen.Index().String().ValuesFunc(func(g *jen.Group) {
			for _, t := range def.RequireServiceReady {
				g.Lit(t)
			}
		})
	}

	if hasConfigValidators {
		values["ConfigValidators"] = jen.Id(entity.String() + configValidatorsFuncSuffix)
	}

	title := entity.Title() + optionsSuffix
	returnValue := jen.Qual(adapterPackage, title).Values(dictFromMap(values, false))
	return jen.Var().Id(title).Op("=").Add(returnValue), nil
}

func validateStateAttributeValues(field, kind string, values []string, prop *Item) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			return fmt.Errorf("%s: %q has duplicate value %q", kind, field, value)
		}
		seen[value] = struct{}{}

		if prop.IsEnum() && !enumContainsString(prop.Enum, value) {
			return fmt.Errorf(
				"%s: %q has enum %v but value %q is not allowed",
				kind,
				field,
				prop.Enum,
				value,
			)
		}
	}
	return nil
}

func stringSliceLiteral(values []string) *jen.Statement {
	items := make([]jen.Code, len(values))
	for i, value := range values {
		items[i] = jen.Lit(value)
	}
	return jen.Index().String().Values(items...)
}

// enumContainsString reports whether enum contains want under fmt.Sprint comparison. Matches the
// runtime semantics of adapter.Equal so generator-time validation and runtime state checks agree.
func enumContainsString(enum []any, want string) bool {
	return slices.ContainsFunc(enum, func(v any) bool { return fmt.Sprint(v) == want })
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
	var errs []error
	view.BlockFunc(func(g *jen.Group) {
		if operation == OperationRead {
			emitReadModifier(g, def)
		}

		if lookupOp != nil {
			var lookupErr error
			g.If(
				jen.Id("d").Dot("Get").Call(jen.Lit(def.DatasourceLookupID())).Op(".").Parens(jen.String()).Op("==").Lit(""),
			).BlockFunc(func(g *jen.Group) {
				lookupErr = genGenericViewOperation(g, 0, 1, def, item, lookupOp)
			})
			if lookupErr != nil {
				errs = append(errs, fmt.Errorf("%s view error: %w", lookupOp.ID, lookupErr))
			}
		}

		for i, op := range operations {
			err := genGenericViewOperation(g, i, len(operations), def, item, op)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s view error: %w", op.ID, err))
			}
		}
	})

	return view, errors.Join(errs...)
}

// emitReadModifier prepends the readModifier call to a Read-style view body.
// No-op when the definition does not opt in via readModifier: true.
func emitReadModifier(g *jen.Group, def *Definition) {
	if !def.ReadModifier {
		return
	}
	g.Err().Op(":=").Id(readModifier).Call(
		jen.Id("ctx"),
		jen.Id("client"),
		jen.Id("d"),
	)
	g.If(jen.Err().Op("!=").Nil()).Block(jen.Return().Err())
}

func genGenericViewOperation(g *jen.Group, funcIndex, funcCount int, def *Definition, item *Item, operation *Operation) error {
	if operation.DatasourceLookup && len(operation.ResultListLookupKeys) == 0 {
		return fmt.Errorf("datasourceLookup operation %q must declare resultListLookupKeys for filtering", operation.ID)
	}

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
			if code := userConfigArg(def, item, inRequest, expandUserConfig); code != nil {
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

	idFields := pathFields(def, item, operation)

	// Some resources might change the ID fields during Update operation,
	// so we need to use get values from the state instead of the plan.
	method := "Get"
	if operation.Type == OperationUpdate && !item.Properties["id"].UsesStateForUnknown() {
		method = "GetState"
	}

	// Gathers parameters: ctx and path parameters to call the client method
	positionalArgs := func(params *jen.Group) {
		params.Id("ctx")
		for _, v := range idFields {
			params.Id("d").Dot(method).Call(jen.Lit(v.Name)).Op(".").Parens(jen.Id(v.GoType()))
		}
		if hasRequest {
			params.Id("req")
		}
	}
	flatten := func(rsp jen.Code) jen.Code {
		v := jen.Id("d").Dot("Flatten").CallFunc(func(g *jen.Group) {
			g.Add(rsp)

			if code := renameFieldsArg(def.Rename, operation.Response, false); code != nil {
				g.Add(code)
			}

			if code := userConfigArg(def, item, inResponse, flattenUserConfig); code != nil {
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

	// flattenMatch wraps the lookup error and flattens the match found by resultListLookupKeys.
	const matchName = "match"
	flattenMatch := func() {
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
					jen.Lit(def.DatasourceLookupID()),
					jen.Id(matchName).Dot(operation.ResultIDField),
				),
				jen.Err().Op("!=").Nil(),
			).Block(jen.Return().Err())
			return
		}

		g.Add(flatten(jen.Op("&").Id(matchName)))
	}

	// Paginated reads fetch the pages with adapter.Paginate, or stop at the match with adapter.PaginateFind
	if pg := operation.Pagination; pg != nil {
		fetch := jen.Id(pageViewName).Call(jen.Id("ctx"), jen.Id("client"), jen.Id("d"))
		if len(operation.ResultListLookupKeys) == 0 {
			itemsVar := rspName + "Items"
			if funcIndex == 0 {
				itemsVar = "items"
			}
			g.List(jen.Id(itemsVar), jen.Err()).Op(":=").Qual(adapterPackage, "Paginate").Call(fetch)
			g.If(jen.Err().Op("!=").Nil()).Block(jen.Return().Err())
			g.Add(flatten(jen.Op("&").Map(jen.String()).Any().Values(jen.Dict{jen.Lit(pg.ItemsKey): jen.Id(itemsVar)})))
			return nil
		}

		found := jen.Id(pageMatchName).Call(jen.Id("d"))
		g.List(jen.Id(matchName), jen.Err()).Op(":=").Qual(adapterPackage, "PaginateFind").Call(fetch, found)
		flattenMatch()
		return nil
	}

	g.Add(clientCall.
		Id("client").Dot(string(operation.ID)).
		CallFunc(func(params *jen.Group) {
			positionalArgs(params)
			for _, k := range sortedKeys(operation.QueryParams) {
				params.Index(jen.Lit(2)).String().Values(jen.Lit(k), jen.Lit(operation.QueryParams[k]))
			}
			if operation.IncludeSecrets {
				params.Qual(commonPackage, "IncludeSecrets").Call()
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
	fieldsMatch, err := genLookupMatch(item, operation, rspCode.Clone().Index(jen.Id("i")))
	if err != nil {
		return err
	}
	g.List(jen.Id(matchName), jen.Err()).Op(":=").
		Qual(adapterPackage, "FindOne").
		CallFunc(func(g *jen.Group) {
			g.Add(rspCode)
			g.Func().Params(jen.Id("i").Int()).Bool().Block(jen.Return(fieldsMatch))
		})
	flattenMatch()
	return nil
}

// pathFields returns the operation's path parameters in the client's argument order.
func pathFields(def *Definition, item *Item, operation *Operation) []*Item {
	inPath := def.Operations.AppearsInID(operation.ID, operation.Type, PathParameter)
	fields := filterAppearsIn(slices.Collect(maps.Values(item.Properties)), inPath)
	sort.Slice(fields, func(i, j int) bool {
		if fields[i].IDAttributePosition != fields[j].IDAttributePosition {
			return fields[i].IDAttributePosition < fields[j].IDAttributePosition
		}
		return fields[i].Name < fields[j].Name
	})
	return fields
}

// genLookupMatch matches the element against the state by resultListLookupKeys,
// e.g. adapter.Equal(rsp[i].Foo, d.Get("foo")) && ...
func genLookupMatch(item *Item, operation *Operation, elem *jen.Statement) (*jen.Statement, error) {
	match := jen.Null()
	for i, key := range sortedKeys(operation.ResultListLookupKeys) {
		fieldName := operation.ResultListLookupKeys[key]
		field, ok := item.Properties[fieldName]
		if !ok {
			return nil, fmt.Errorf("unknown lookup key %q in result list for operation %q", fieldName, operation.ID)
		}
		if i > 0 {
			match.Op("&&")
		}
		match.Qual(adapterPackage, "Equal").Call(elem.Clone().Dot(key), jen.Id("d").Dot("Get").Call(jen.Lit(field.Name)))
	}
	return match, nil
}

// pageViewName returns the fetch for adapter.Paginate, pageMatchName returns the predicate for
// adapter.PaginateFind. Both are shared by readView and ConflictOptions.Load.
const (
	pageViewName  = "pageView"
	pageMatchName = "pageMatch"
)

// pagedOperation returns the paginated read operation with a view, at most one is supported.
func pagedOperation(def *Definition) (*Operation, error) {
	var found *Operation
	for _, op := range def.Operations {
		if op.Pagination == nil || op.DisableView {
			continue
		}
		if found != nil {
			return nil, fmt.Errorf("only one paginated operation is supported, got %q and %q", found.ID, op.ID)
		}
		found = op
	}
	return found, nil
}

// genConflictOptions renders adapter.ConflictOptions that list the scope with pageView and match it with pageMatch.
// The scope is the paginated read's path, the keys are its resultListLookupKeys.
func genConflictOptions(def *Definition, item *Item) (jen.Code, error) {
	op, err := pagedOperation(def)
	if err != nil {
		return nil, err
	}
	if op == nil || len(op.ResultListLookupKeys) == 0 {
		return nil, errors.New("requires a read operation with pagination and resultListLookupKeys")
	}

	keys := lo.Map(sortedKeys(op.ResultListLookupKeys), func(k string, _ int) string { return op.ResultListLookupKeys[k] })
	parent := lo.Map(pathFields(def, item, op), func(v *Item, _ int) string { return v.Name })
	return jen.Op("&").Qual(adapterPackage, "ConflictOptions").Values(jen.Dict{
		jen.Id("Parent"): stringSliceLiteral(parent),
		jen.Id("Keys"):   stringSliceLiteral(keys),
		jen.Id("Load"):   jen.Qual(adapterPackage, "ConflictList").Call(jen.Id(pageViewName), jen.Id(pageMatchName)),
	}), nil
}

// genPageMatch renders the pageMatch function:
//
//	func pageMatch(d) func(item T) bool
func genPageMatch(def *Definition, item *Item, operation *Operation) (jen.Code, error) {
	match, err := genLookupMatch(item, operation, jen.Id("item"))
	if err != nil {
		return nil, err
	}
	predicate := func() *jen.Statement {
		return jen.Func().Params(jen.Id("item").Add(pageItemType(def, operation.Pagination))).Bool()
	}
	return jen.Func().Id(pageMatchName).
		Params(jen.Id("d").Qual(adapterPackage, resourceData)).
		Add(predicate()).
		Block(jen.Return(predicate().Block(jen.Return(match)))), nil
}

// pageItemType is the client's list item type, which trims "s", e.g. BillingGroups -> BillingGroupOut.
func pageItemType(def *Definition, pg *Pagination) *jen.Statement {
	return jen.Qual(avnGenHandlerPackage+def.ClientHandler, strings.TrimSuffix(pascalCase(pg.ItemsKey), "s")+"Out")
}

// genPageView renders the pageView function:
//
//	func pageView(ctx, client, d) func(cursor string) ([]T, *string, error)
func genPageView(def *Definition, item *Item, operation *Operation) jen.Code {
	positionalArgs := func(params *jen.Group) {
		params.Id("ctx")
		for _, v := range pathFields(def, item, operation) {
			params.Id("d").Dot("Get").Call(jen.Lit(v.Name)).Op(".").Parens(jen.Id(v.GoType()))
		}
	}
	fetch, itemType := genPageFetch(def, operation, positionalArgs)
	return jen.Func().Id(pageViewName).
		Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("client").Qual(avnGenPackage, "Client"),
			jen.Id("d").Qual(adapterPackage, resourceData),
		).
		Func().Params(jen.Id("cursor").String()).Params(jen.Index().Add(itemType), jen.Op("*").String(), jen.Error()).
		Block(jen.Return(fetch))
}

// genPageFetch renders the adapter.Paginate fetch function and the client's list item type.
func genPageFetch(def *Definition, operation *Operation, positionalArgs func(*jen.Group)) (fetch, itemType *jen.Statement) {
	pg := operation.Pagination
	query := jen.Qual(adapterPackage, "QueryParams").CallFunc(func(params *jen.Group) {
		if pg.PageSizeParam != "" {
			params.Lit(pg.PageSizeParam)
			params.Lit(strconv.Itoa(pg.PageSize))
		}
		for _, k := range sortedKeys(operation.QueryParams) {
			params.Lit(k)
			params.Lit(operation.QueryParams[k])
		}
		params.Lit(pg.CursorParam)
		params.Id("cursor")
	})
	if operation.IncludeSecrets {
		query = jen.Append(query, jen.Qual(commonPackage, "IncludeSecrets").Call())
	}

	itemType = pageItemType(def, pg)
	fetch = jen.Func().Params(jen.Id("cursor").String()).Params(jen.Index().Add(itemType.Clone()), jen.Op("*").String(), jen.Error()).Block(
		jen.List(jen.Id("rsp"), jen.Err()).Op(":=").Id("client").Dot(string(operation.ID)).CallFunc(func(params *jen.Group) {
			positionalArgs(params)
			params.Add(query).Op("...")
		}),
		jen.If(jen.Err().Op("!=").Nil()).Block(jen.Return(jen.Nil(), jen.Nil(), jen.Err())),
		jen.Return(jen.Id("rsp").Dot(pascalCase(pg.ItemsKey)), jen.Id("rsp").Dot(pascalCase(pg.NextField)), jen.Nil()),
	)
	return fetch, itemType
}

// lookupFieldsList renders an op's ResultListLookupKeys tf-attribute values as a
// backticked human-readable list joined by " and " (or " or " when OR mode is set).
func lookupFieldsList(op *Operation) string {
	values := lo.Values(op.ResultListLookupKeys)
	slices.Sort(values)
	return humanizeCodeList(values, " and ")
}

// datasourceLookupValidators emits the datasource's ConfigValidators function.
// When `datasource.exactlyOneOf` is set, it emits a single
// `datasourcevalidator.ExactlyOneOf(...)` over its members; otherwise it
// falls back to the lookup-contract pair (id vs composedOf[0]) plus
// RequiredTogether when composedOf has more than one field.
func datasourceLookupValidators(def *Definition) jen.Code {
	pkg := datasourceType.Import(entityValidatorPkgFmt)
	matchRoots := func(names ...string) []jen.Code {
		return lo.Map(names, func(name string, _ int) jen.Code {
			return jen.Qual(pathPackage, "MatchRoot").Call(jen.Lit(name))
		})
	}

	var names, composedOf []string
	if eo := def.Datasource.ExactlyOneOf; len(eo) > 0 {
		names = eo
	} else if def.DatasourceLookupOp() != nil {
		composedOf = def.DatasourceLookupComposedOf()
		names = []string{def.DatasourceLookupID(), composedOf[0]}
	}

	var validators []jen.Code
	if len(names) > 0 {
		validators = append(validators, jen.Qual(pkg, "ExactlyOneOf").Custom(multilineCall(), matchRoots(names...)...))
	}
	if len(composedOf) > 1 {
		validators = append(validators, jen.Qual(pkg, "RequiredTogether").Custom(multilineCall(), matchRoots(composedOf...)...))
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

func identityFieldNames(root *Item) []string {
	fields := root.GetIDFields()
	names := make([]string, len(fields))
	for i, f := range fields {
		names[i] = f.Name
	}
	return names
}

func identityFieldsMutable(def *Definition, item *Item) bool {
	for _, names := range [][]string{def.IdentityFields, def.IdentityFieldsOptional} {
		for _, name := range names {
			prop, ok := item.Properties[name]
			if !ok {
				continue
			}
			if def.Operations.AppearsInHandler(prop.AppearsIn, OperationUpdate, RequestBody) {
				return true
			}
		}
	}
	return false
}
