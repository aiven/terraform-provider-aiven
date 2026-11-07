package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func paginatedOpenAPI() *OpenAPIDoc {
	return paginatedOpenAPIWith(paginatedOpenAPIOpts{
		cursorName: "cursor",
		pageName:   "limit",
		pageSchema: &OASchema{Type: SchemaTypeInteger, Maximum: "999"},
	})
}

type paginatedOpenAPIOpts struct {
	cursorName string
	pageName   string
	pageSchema *OASchema
}

func paginatedOpenAPIWith(o paginatedOpenAPIOpts) *OpenAPIDoc {
	doc := &OpenAPIDoc{
		Paths: map[string]map[string]*OAPath{"/x": {"get": {
			OperationID: "List",
			Parameters: []*OAParameter{
				{Ref: "#/components/parameters/" + o.cursorName},
				{Ref: "#/components/parameters/" + o.pageName},
			},
		}}},
	}
	doc.Components.Parameters = map[string]*OAParameter{
		o.cursorName: {In: "query", Name: o.cursorName, Schema: &OASchema{Type: SchemaTypeString}},
		o.pageName:   {In: "query", Name: o.pageName, Schema: o.pageSchema},
	}
	doc.Components.Schemas = map[string]*OASchema{"Out": {
		Type: SchemaTypeObject,
		Properties: map[string]*OASchema{
			"billing_groups": {Type: SchemaTypeArray, Items: &OASchema{Type: SchemaTypeString}},
			"next":           {Type: SchemaTypeString},
		},
	}}
	return doc
}

func TestValidatePagination(t *testing.T) {
	t.Parallel()

	baseOp := func() *Operation {
		return &Operation{
			ID:   "List",
			Type: OperationRead,
			Pagination: &Pagination{
				NextField:     "next",
				CursorParam:   "cursor",
				PageSizeParam: "limit",
				PageSize:      999,
			},
		}
	}

	cases := []struct {
		name         string
		doc          func() *OpenAPIDoc
		modify       func(doc *OpenAPIDoc, op *Operation)
		remove       []string
		wantPageSize int
		wantErr      string
	}{
		{
			name:         "ok paginated default",
			doc:          paginatedOpenAPI,
			wantPageSize: 999,
		},
		{
			name: "ok pageSize defaults to the default",
			doc: func() *OpenAPIDoc {
				return paginatedOpenAPIWith(paginatedOpenAPIOpts{
					cursorName: "after",
					pageName:   "maxItems",
					pageSchema: &OASchema{Type: SchemaTypeInteger, Default: float64(250)},
				})
			},
			modify: func(doc *OpenAPIDoc, op *Operation) {
				doc.Components.Schemas["Out"].Properties = map[string]*OASchema{
					"databases": {Type: SchemaTypeArray, Items: &OASchema{Type: SchemaTypeString}},
					"next":      {Type: SchemaTypeString},
				}
				op.Pagination = &Pagination{
					NextField:     "next",
					CursorParam:   "after",
					PageSizeParam: "maxItems",
				}
			},
			wantPageSize: 250,
		},
		{
			name: "ok no page size",
			doc:  paginatedOpenAPI,
			modify: func(doc *OpenAPIDoc, op *Operation) {
				doc.Paths["/x"]["get"].Parameters = doc.Paths["/x"]["get"].Parameters[:1]
				op.Pagination.PageSizeParam = ""
				op.Pagination.PageSize = 0
			},
		},
		{
			name: "ok not paginated nil pagination no next",
			doc:  paginatedOpenAPI,
			modify: func(doc *OpenAPIDoc, op *Operation) {
				op.Pagination = nil
				doc.Components.Schemas["Out"].Properties = map[string]*OASchema{
					"billing_groups": {Type: SchemaTypeArray, Items: &OASchema{Type: SchemaTypeString}},
				}
			},
		},
		{
			name: "ok matching resultKey",
			doc:  paginatedOpenAPI,
			modify: func(doc *OpenAPIDoc, op *Operation) {
				doc.Components.Schemas["Out"].Properties = map[string]*OASchema{
					"databases": {Type: SchemaTypeArray, Items: &OASchema{Type: SchemaTypeString}},
					"next":      {Type: SchemaTypeString},
				}
				op.ResultKey = "databases"
			},
		},
		{
			name: "ok prev in remove",
			doc:  paginatedOpenAPI,
			modify: func(doc *OpenAPIDoc, _ *Operation) {
				doc.Components.Schemas["Out"].Properties["prev"] = &OASchema{Type: SchemaTypeString}
				doc.Components.Schemas["Out"].Required = []string{"next", "prev"}
			},
			remove: []string{"prev"},
		},
		{
			name: "err prev not in remove",
			doc:  paginatedOpenAPI,
			modify: func(doc *OpenAPIDoc, _ *Operation) {
				doc.Components.Schemas["Out"].Properties["prev"] = &OASchema{Type: SchemaTypeString}
			},
			wantErr: `want one property besides next and the remove list, got ["billing_groups" "prev"]`,
		},
		{
			name: "ok custom nextField",
			doc:  paginatedOpenAPI,
			modify: func(doc *OpenAPIDoc, op *Operation) {
				out := doc.Components.Schemas["Out"]
				out.Properties["next_page"] = out.Properties["next"]
				op.Pagination.NextField = "next_page"
			},
			remove: []string{"next"},
		},
		{
			name: "err empty nextField",
			doc:  paginatedOpenAPI,
			modify: func(_ *OpenAPIDoc, op *Operation) {
				op.Pagination.NextField = ""
			},
			wantErr: "nextField is required",
		},
		{
			name: "err nextField not a string",
			doc:  paginatedOpenAPI,
			modify: func(doc *OpenAPIDoc, _ *Operation) {
				doc.Components.Schemas["Out"].Properties["next"] = &OASchema{Type: SchemaTypeInteger}
			},
			wantErr: `"next" must be a string, got "integer"`,
		},
		{
			name: "err not a read",
			doc:  paginatedOpenAPI,
			modify: func(_ *OpenAPIDoc, op *Operation) {
				op.Type = OperationUpdate
			},
			wantErr: "update operation, want read",
		},
		{
			name: "err response without next",
			doc:  paginatedOpenAPI,
			modify: func(doc *OpenAPIDoc, _ *Operation) {
				doc.Components.Schemas["Out"].Properties = map[string]*OASchema{
					"billing_groups": {Type: SchemaTypeArray, Items: &OASchema{Type: SchemaTypeString}},
				}
			},
			wantErr: `the response has no "next" field`,
		},
		{
			name: "err empty cursorParam",
			doc:  paginatedOpenAPI,
			modify: func(_ *OpenAPIDoc, op *Operation) {
				op.Pagination.CursorParam = ""
			},
			wantErr: "cursorParam is required",
		},
		{
			name: "err unknown cursor param",
			doc:  paginatedOpenAPI,
			modify: func(doc *OpenAPIDoc, _ *Operation) {
				doc.Paths["/x"]["get"].Parameters = doc.Paths["/x"]["get"].Parameters[1:]
			},
			wantErr: `no "cursor" query parameter`,
		},
		{
			name: "ok pageSize defaults to the maximum",
			doc:  paginatedOpenAPI,
			modify: func(_ *OpenAPIDoc, op *Operation) {
				op.Pagination.PageSize = 0
			},
			wantPageSize: 999,
		},
		{
			name: "err no maximum or default",
			doc:  paginatedOpenAPI,
			modify: func(doc *OpenAPIDoc, op *Operation) {
				doc.Components.Parameters["limit"].Schema = &OASchema{Type: SchemaTypeInteger}
				op.Pagination.PageSize = 0
			},
			wantErr: "limit has no maximum or default, set pageSize",
		},
		{
			name: "err pageSize under minimum",
			doc:  paginatedOpenAPI,
			modify: func(doc *OpenAPIDoc, op *Operation) {
				doc.Components.Parameters["limit"].Schema = &OASchema{Type: SchemaTypeInteger, Minimum: "30", Maximum: "999"}
				op.Pagination.PageSize = 10
			},
			wantErr: "pageSize 10 is under the limit minimum 30",
		},
		{
			name: "err pageSize without pageSizeParam",
			doc:  paginatedOpenAPI,
			modify: func(_ *OpenAPIDoc, op *Operation) {
				op.Pagination.PageSizeParam = ""
			},
			wantErr: "pageSize needs pageSizeParam",
		},
		{
			name: "err negative pageSize",
			doc:  paginatedOpenAPI,
			modify: func(_ *OpenAPIDoc, op *Operation) {
				op.Pagination.PageSize = -1
			},
			wantErr: "pageSize must be positive",
		},
		{
			name: "err unknown page size param",
			doc:  paginatedOpenAPI,
			modify: func(doc *OpenAPIDoc, _ *Operation) {
				doc.Paths["/x"]["get"].Parameters = doc.Paths["/x"]["get"].Parameters[:1]
			},
			wantErr: `no "limit" query parameter`,
		},
		{
			name: "err pageSize over maximum",
			doc:  paginatedOpenAPI,
			modify: func(_ *OpenAPIDoc, op *Operation) {
				op.Pagination.PageSize = 1000
			},
			wantErr: "pageSize 1000 is over the limit maximum 999",
		},
		{
			name: "err another property besides nav",
			doc:  paginatedOpenAPI,
			modify: func(doc *OpenAPIDoc, _ *Operation) {
				doc.Components.Schemas["Out"].Properties["total"] = &OASchema{Type: SchemaTypeInteger}
			},
			wantErr: `want one property besides next and the remove list, got ["billing_groups" "total"]`,
		},
		{
			name: "err items not array",
			doc:  paginatedOpenAPI,
			modify: func(doc *OpenAPIDoc, _ *Operation) {
				doc.Components.Schemas["Out"].Properties["billing_groups"] = &OASchema{Type: SchemaTypeObject}
			},
			wantErr: `"billing_groups" must be an array`,
		},
		{
			name: "err resultToKey",
			doc:  paginatedOpenAPI,
			modify: func(_ *OpenAPIDoc, op *Operation) {
				op.ResultToKey = "billing_groups"
			},
			wantErr: "remove resultKeyField and resultToKey",
		},
		{
			name: "err mismatched resultKey",
			doc:  paginatedOpenAPI,
			modify: func(doc *OpenAPIDoc, op *Operation) {
				doc.Components.Schemas["Out"].Properties = map[string]*OASchema{
					"databases": {Type: SchemaTypeArray, Items: &OASchema{Type: SchemaTypeString}},
					"next":      {Type: SchemaTypeString},
				}
				op.ResultKey = "items"
			},
			wantErr: `pagination: resultKey "items" must be "databases" or omitted`,
		},
		{
			name: "err resultKeyField",
			doc:  paginatedOpenAPI,
			modify: func(doc *OpenAPIDoc, op *Operation) {
				doc.Components.Schemas["Out"].Properties = map[string]*OASchema{
					"databases": {Type: SchemaTypeArray, Items: &OASchema{Type: SchemaTypeString}},
					"next":      {Type: SchemaTypeString},
				}
				op.ResultKeyField = "Databases"
			},
			wantErr: "remove resultKeyField and resultToKey",
		},
		{
			name: "err cursor in queryParams",
			doc:  paginatedOpenAPI,
			modify: func(_ *OpenAPIDoc, op *Operation) {
				op.QueryParams = map[string]string{"cursor": "x"}
			},
			wantErr: `queryParams can't set "cursor"`,
		},
		{
			name: "err maxItems in queryParams",
			doc: func() *OpenAPIDoc {
				return paginatedOpenAPIWith(paginatedOpenAPIOpts{
					cursorName: "after",
					pageName:   "maxItems",
					pageSchema: &OASchema{Type: SchemaTypeInteger, Default: float64(250)},
				})
			},
			modify: func(doc *OpenAPIDoc, op *Operation) {
				doc.Components.Schemas["Out"].Properties = map[string]*OASchema{
					"databases": {Type: SchemaTypeArray, Items: &OASchema{Type: SchemaTypeString}},
					"next":      {Type: SchemaTypeString},
				}
				op.Pagination = &Pagination{
					NextField:     "next",
					CursorParam:   "after",
					PageSizeParam: "maxItems",
					PageSize:      250,
				}
				op.QueryParams = map[string]string{"maxItems": "10"}
			},
			wantErr: `queryParams can't set "maxItems"`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			doc := tc.doc()
			op := baseOp()
			if tc.modify != nil {
				tc.modify(doc, op)
			}
			response := doc.Components.Schemas["Out"]

			got, err := resolvePagination(&Scope{OpenAPI: doc, Definition: &Definition{Remove: tc.remove}}, doc.Paths["/x"]["get"], op, response)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			if op.Pagination == nil {
				require.Same(t, response, got)
				return
			}
			require.Contains(t, response.Properties, "next", "the shared schema must keep next")
			require.Equal(t, []string{op.Pagination.ItemsKey}, sortedKeys(got.Properties))
			require.Empty(t, got.Required)
			if tc.wantPageSize != 0 {
				require.Equal(t, tc.wantPageSize, op.Pagination.PageSize)
			}
		})
	}
}

func TestValidateResultToKey(t *testing.T) {
	t.Parallel()

	op := &Operation{ResultToKey: "users"}
	single := &OASchema{Properties: map[string]*OASchema{"users": {}}}
	require.NoError(t, validateResultToKey(op, single))

	object := &OASchema{Properties: map[string]*OASchema{"users": {}, "next": {}}}
	require.ErrorContains(t, validateResultToKey(op, object), "the response has 2 properties")
}

func pagedDefinition() *Definition {
	return &Definition{Operations: []*Operation{{
		ID:                   "ServiceDatabaseList",
		Type:                 OperationRead,
		ResultListLookupKeys: map[string]string{"DatabaseName": "database_name"},
		Pagination:           &Pagination{ItemsKey: "databases", NextField: "next", CursorParam: "after"},
	}}}
}

func TestGenConflictOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		setup   func(def *Definition)
		wantErr string
	}{
		{
			name:    "no pagination",
			setup:   func(def *Definition) { def.Operations[0].Pagination = nil },
			wantErr: "requires a read operation with pagination and resultListLookupKeys",
		},
		{
			name:    "disabled view",
			setup:   func(def *Definition) { def.Operations[0].DisableView = true },
			wantErr: "requires a read operation with pagination and resultListLookupKeys",
		},
		{
			name:    "no lookup keys",
			setup:   func(def *Definition) { def.Operations[0].ResultListLookupKeys = nil },
			wantErr: "requires a read operation with pagination and resultListLookupKeys",
		},
		{
			name: "two paginated operations",
			setup: func(def *Definition) {
				dup := *def.Operations[0]
				dup.ID = "ServiceDatabaseList2"
				def.Operations = append(def.Operations, &dup)
			},
			wantErr: "only one paginated operation is supported",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			def := pagedDefinition()
			tc.setup(def)
			_, err := genConflictOptions(def, &Item{})
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestGenPageMatch_UnknownLookupAttribute(t *testing.T) {
	t.Parallel()

	def := pagedDefinition()
	_, err := genPageMatch(def, &Item{}, def.Operations[0])
	require.ErrorContains(t, err, "unknown lookup key")
}
