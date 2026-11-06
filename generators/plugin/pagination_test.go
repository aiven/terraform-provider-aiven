package main

import (
	"strings"
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

func TestGenGenericView_CursorPaginationPointerNext(t *testing.T) {
	t.Parallel()

	org := &Item{
		Name:                "organization_id",
		Type:                SchemaTypeString,
		Required:            true,
		IDAttribute:         true,
		IDAttributePosition: 0,
		AppearsIn:           Operations{{ID: "OrgList", Type: OperationRead}}.AppearsInID("OrgList", OperationRead, PathParameter),
	}
	bg := &Item{
		Name:      "billing_groups",
		Type:      SchemaTypeArrayOrdered,
		Computed:  true,
		AppearsIn: Operations{{ID: "OrgList", Type: OperationRead}}.AppearsInID("OrgList", OperationRead, ResponseBody),
	}
	item := &Item{Properties: map[string]*Item{"organization_id": org, "billing_groups": bg}}
	setParents(item, nil)

	def := &Definition{
		typeName:      "aiven_organization_billing_group_list",
		Datasource:    &SchemaMeta{},
		ClientHandler: "organizationbilling",
		Operations: []*Operation{{
			ID:   "OrgList",
			Type: OperationRead,
			Pagination: &Pagination{
				ItemsKey:      "billing_groups",
				NextField:     "next",
				CursorParam:   "cursor",
				PageSizeParam: "limit",
				PageSize:      999,
			},
			Response: &OASchema{Properties: map[string]*OASchema{"billing_groups": {}}},
		}},
	}

	code, err := genGenericView(def, item, OperationRead)
	require.NoError(t, err)
	got := renderCode(t, code)

	want := `package projectvpc

import (
	"context"
	goclientcodegen "github.com/aiven/go-client-codegen"
	organizationbilling "github.com/aiven/go-client-codegen/handler/organizationbilling"
	adapter "github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func readView(ctx context.Context, client goclientcodegen.Client, d adapter.ResourceData) error {
	items, err := adapter.Paginate(func(cursor string) ([]organizationbilling.BillingGroupOut, *string, error) {
		rsp, err := client.OrgList(ctx, d.Get("organization_id").(string), adapter.QueryParams("limit", "999", "cursor", cursor)...)
		if err != nil {
			return nil, nil, err
		}
		return rsp.BillingGroups, rsp.Next, nil
	})
	if err != nil {
		return err
	}
	return d.Flatten(&map[string]any{"billing_groups": items})
}
`
	require.Equal(t, strings.TrimSpace(want), strings.TrimSpace(got))
}

func TestGenGenericView_CursorPaginationFindOne(t *testing.T) {
	t.Parallel()

	project := &Item{
		Name:                "project",
		Type:                SchemaTypeString,
		Required:            true,
		IDAttribute:         true,
		IDAttributePosition: 0,
		AppearsIn:           Operations{{ID: "ServiceDatabaseList", Type: OperationRead}}.AppearsInID("ServiceDatabaseList", OperationRead, PathParameter),
	}
	svc := &Item{
		Name:                "service_name",
		Type:                SchemaTypeString,
		Required:            true,
		IDAttribute:         true,
		IDAttributePosition: 1,
		AppearsIn:           project.AppearsIn,
	}
	dbName := &Item{
		Name:      "database_name",
		Type:      SchemaTypeString,
		Required:  true,
		AppearsIn: Operations{{ID: "ServiceDatabaseList", Type: OperationRead}}.AppearsInID("ServiceDatabaseList", OperationRead, ResponseBody),
	}
	lc := &Item{
		Name:      "lc_collate",
		Type:      SchemaTypeString,
		Computed:  true,
		AppearsIn: dbName.AppearsIn,
	}
	root := &Item{Properties: map[string]*Item{
		"project": project, "service_name": svc, "database_name": dbName, "lc_collate": lc,
	}}
	setParents(root, nil)

	def := &Definition{
		typeName:      "aiven_pg_database",
		ClientHandler: "service",
		Resource:      &SchemaMeta{},
		Datasource:    &SchemaMeta{},
		Operations: []*Operation{{
			ID:                   "ServiceDatabaseList",
			Type:                 OperationRead,
			ResultKey:            "databases",
			ResultListLookupKeys: map[string]string{"DatabaseName": "database_name"},
			Pagination: &Pagination{
				ItemsKey:      "databases",
				NextField:     "next",
				CursorParam:   "after",
				PageSizeParam: "maxItems",
				PageSize:      250,
			},
			Response: &OASchema{Properties: map[string]*OASchema{
				"databases": {},
			}},
		}},
	}

	code, err := genGenericView(def, root, OperationRead)
	require.NoError(t, err)
	got := renderCode(t, code)

	want := "package projectvpc\n\n" +
		"import (\n" +
		"\t\"context\"\n" +
		"\t\"fmt\"\n" +
		"\tgoclientcodegen \"github.com/aiven/go-client-codegen\"\n" +
		"\tservice \"github.com/aiven/go-client-codegen/handler/service\"\n" +
		"\tadapter \"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter\"\n" +
		")\n\n" +
		"func readView(ctx context.Context, client goclientcodegen.Client, d adapter.ResourceData) error {\n" +
		"\tmatch, err := adapter.PaginateFind(func(cursor string) ([]service.DatabaseOut, *string, error) {\n" +
		"\t\trsp, err := client.ServiceDatabaseList(ctx, d.Get(\"project\").(string), d.Get(\"service_name\").(string), adapter.QueryParams(\"maxItems\", \"250\", \"after\", cursor)...)\n" +
		"\t\tif err != nil {\n" +
		"\t\t\treturn nil, nil, err\n" +
		"\t\t}\n" +
		"\t\treturn rsp.Databases, rsp.Next, nil\n" +
		"\t}, func(item service.DatabaseOut) bool {\n" +
		"\t\treturn adapter.Equal(item.DatabaseName, d.Get(\"database_name\"))\n" +
		"\t})\n" +

		"\tif err != nil {\n" +
		"\t\treturn fmt.Errorf(\"lookup `aiven_pg_database` by `database_name`: %w\", err)\n" +
		"\t}\n" +
		"\treturn d.Flatten(&match)\n" +
		"}\n"
	require.Equal(t, strings.TrimSpace(want), strings.TrimSpace(got))
}

func TestGenGenericView_CursorPaginationQueryParams(t *testing.T) {
	t.Parallel()

	org := &Item{
		Name:                "organization_id",
		Type:                SchemaTypeString,
		Required:            true,
		IDAttribute:         true,
		IDAttributePosition: 0,
		AppearsIn:           Operations{{ID: "OrgList", Type: OperationRead}}.AppearsInID("OrgList", OperationRead, PathParameter),
	}
	bg := &Item{
		Name:      "billing_groups",
		Type:      SchemaTypeArrayOrdered,
		Computed:  true,
		AppearsIn: Operations{{ID: "OrgList", Type: OperationRead}}.AppearsInID("OrgList", OperationRead, ResponseBody),
	}
	item := &Item{Properties: map[string]*Item{"organization_id": org, "billing_groups": bg}}
	setParents(item, nil)

	def := &Definition{
		typeName:      "aiven_organization_billing_group_list",
		Datasource:    &SchemaMeta{},
		ClientHandler: "organizationbilling",
		Operations: []*Operation{{
			ID:             "OrgList",
			Type:           OperationRead,
			IncludeSecrets: true,
			QueryParams:    map[string]string{"expand": "all"},
			Pagination: &Pagination{
				ItemsKey:      "billing_groups",
				NextField:     "next",
				CursorParam:   "cursor",
				PageSizeParam: "limit",
				PageSize:      999,
			},
			Response: &OASchema{Properties: map[string]*OASchema{"billing_groups": {}}},
		}},
	}

	code, err := genGenericView(def, item, OperationRead)
	require.NoError(t, err)
	got := renderCode(t, code)
	require.Contains(t, got, `append(adapter.QueryParams("limit", "999", "expand", "all", "cursor", cursor), common.IncludeSecrets())...`)

	pg := def.Operations[0].Pagination
	pg.PageSizeParam, pg.PageSize = "", 0
	code, err = genGenericView(def, item, OperationRead)
	require.NoError(t, err)
	got = renderCode(t, code)
	require.Contains(t, got, `append(adapter.QueryParams("expand", "all", "cursor", cursor), common.IncludeSecrets())...`)
}
