package main

import (
	"fmt"
	"maps"
	"slices"
)

// queryParam resolves a query parameter by name; (nil, nil) if absent.
func queryParam(scope *Scope, path *OAPath, name string) (*OAParameter, error) {
	for _, param := range path.Parameters {
		p, err := scope.OpenAPI.getParameterRef(param.Ref)
		if err != nil {
			return nil, err
		}
		if p.In == "query" && p.Name == name {
			return p, nil
		}
	}
	return nil, nil
}

// withoutPaginationFields drops nextField and the `remove` list from a copy, so the shared component keeps them.
func withoutPaginationFields(def *Definition, s *OASchema, nextField string) (*OASchema, error) {
	cp := *s
	cp.Properties = maps.Clone(s.Properties)
	for name := range s.Properties {
		removed, err := def.IsRemoved(name)
		if err != nil {
			return nil, err
		}
		if removed || name == nextField {
			delete(cp.Properties, name)
		}
	}
	cp.Required = slices.DeleteFunc(slices.Clone(s.Required), func(name string) bool {
		return cp.Properties[name] == nil
	})
	return &cp, nil
}

// resolvePageSize defaults pageSize to the parameter's maximum, else its default, and checks the range.
func resolvePageSize(pg *Pagination, s *OASchema) error {
	if pg.PageSize == 0 {
		if s.Maximum != "" {
			pg.PageSize = numberToInt(s.Maximum, false)
		} else if v, ok := s.Default.(float64); ok {
			pg.PageSize = int(v)
		} else {
			return fmt.Errorf("pagination: %s has no maximum or default, set pageSize", pg.PageSizeParam)
		}
	}
	if s.Maximum != "" && pg.PageSize > numberToInt(s.Maximum, false) {
		return fmt.Errorf("pagination: pageSize %d is over the %s maximum %s", pg.PageSize, pg.PageSizeParam, s.Maximum)
	}
	if s.Minimum != "" && pg.PageSize < numberToInt(s.Minimum, true) {
		return fmt.Errorf("pagination: pageSize %d is under the %s minimum %s", pg.PageSize, pg.PageSizeParam, s.Minimum)
	}
	return nil
}

// resolvePagination checks the operation's pagination against the spec and sets its ItemsKey.
// Returns the response schema without pagination fields, or the response as is when not paginated.
func resolvePagination(scope *Scope, path *OAPath, operation *Operation, response *OASchema) (*OASchema, error) {
	pg := operation.Pagination
	if pg == nil {
		return response, nil
	}
	if operation.Type != OperationRead {
		return nil, fmt.Errorf("pagination: %s operation, want read", operation.Type)
	}
	if pg.NextField == "" {
		return nil, fmt.Errorf("pagination: nextField is required")
	}
	nextSchema, err := scope.OpenAPI.resolveRef(response.Properties[pg.NextField])
	if err != nil {
		return nil, fmt.Errorf("pagination: %s: %w", pg.NextField, err)
	}
	if nextSchema == nil {
		return nil, fmt.Errorf("pagination: the response has no %q field", pg.NextField)
	}
	if nextSchema.Type != SchemaTypeString {
		return nil, fmt.Errorf("pagination: %q must be a string, got %q", pg.NextField, nextSchema.Type)
	}

	if pg.CursorParam == "" {
		return nil, fmt.Errorf("pagination: cursorParam is required")
	}
	cursorParam, err := queryParam(scope, path, pg.CursorParam)
	if err != nil {
		return nil, err
	}
	if cursorParam == nil {
		return nil, fmt.Errorf("pagination: no %q query parameter", pg.CursorParam)
	}

	if pg.PageSize < 0 {
		return nil, fmt.Errorf("pagination: pageSize must be positive")
	}
	if pg.PageSizeParam == "" && pg.PageSize != 0 {
		return nil, fmt.Errorf("pagination: pageSize needs pageSizeParam")
	}
	if pg.PageSizeParam != "" {
		pageParam, err := queryParam(scope, path, pg.PageSizeParam)
		if err != nil {
			return nil, err
		}
		if pageParam == nil || pageParam.Schema == nil {
			return nil, fmt.Errorf("pagination: no %q query parameter", pg.PageSizeParam)
		}
		pageSchema, err := scope.OpenAPI.resolveRef(pageParam.Schema)
		if err != nil {
			return nil, fmt.Errorf("pagination: %s: %w", pg.PageSizeParam, err)
		}
		err = resolvePageSize(pg, pageSchema)
		if err != nil {
			return nil, err
		}
	}

	stripped, err := withoutPaginationFields(scope.Definition, response, pg.NextField)
	if err != nil {
		return nil, err
	}
	if len(stripped.Properties) != 1 {
		return nil, fmt.Errorf("pagination: want one property besides %s and the remove list, got %q", pg.NextField, sortedKeys(stripped.Properties))
	}
	pg.ItemsKey = sortedKeys(stripped.Properties)[0]
	itemsSchema, err := scope.OpenAPI.resolveRef(stripped.Properties[pg.ItemsKey])
	if err != nil {
		return nil, fmt.Errorf("pagination: %s: %w", pg.ItemsKey, err)
	}
	if itemsSchema.Type != SchemaTypeArray && itemsSchema.Type != SchemaTypeArrayOrdered {
		return nil, fmt.Errorf("pagination: %q must be an array, got %q", pg.ItemsKey, itemsSchema.Type)
	}

	if operation.ResultKey != "" && operation.ResultKey != pg.ItemsKey {
		return nil, fmt.Errorf("pagination: resultKey %q must be %q or omitted", operation.ResultKey, pg.ItemsKey)
	}
	if operation.ResultKeyField != "" || operation.ResultToKey != "" {
		return nil, fmt.Errorf("pagination: remove resultKeyField and resultToKey, the pages are merged by %q", pg.ItemsKey)
	}

	for k := range operation.QueryParams {
		if k == pg.CursorParam || k == pg.PageSizeParam {
			return nil, fmt.Errorf("pagination: queryParams can't set %q", k)
		}
	}
	return stripped, nil
}
