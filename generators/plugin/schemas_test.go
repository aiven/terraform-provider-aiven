package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// rootWithObject builds a resource root that holds a single object property.
func rootWithObject(object *Item) *Item {
	root := &Item{
		Name:       "root",
		Type:       SchemaTypeObject,
		Properties: map[string]*Item{object.Name: object},
	}

	object.Parent = root
	for _, v := range object.Properties {
		v.Parent = object
	}
	return root
}

func genResourceSchemaCode(t *testing.T, root *Item) string {
	t.Helper()

	code, err := genSchema(&Definition{Resource: &SchemaMeta{}}, resourceType, root)
	require.NoError(t, err)
	return renderCode(t, code)
}

// A read-only collection renders as a computed attribute, see Item.RendersAsAttribute.
func TestGenSchemaRendersReadOnlyObjectAsAttribute(t *testing.T) {
	got := genResourceSchemaCode(t, rootWithObject(&Item{
		Name:     "current_deployment",
		Type:     SchemaTypeObject,
		Computed: true,
		Properties: map[string]*Item{
			"status": {Name: "status", Type: SchemaTypeString, Computed: true},
		},
	}))

	require.Contains(t, got, `"current_deployment": schema.ListNestedAttribute{`)
	// Validators only see the configuration.
	require.NotContains(t, got, "listvalidator")
}

// Objects and ordered lists take different paths through the internal schema generator.
func TestGenSchemaInternalMarksReadOnlyCollectionsAsAttributes(t *testing.T) {
	root := &Item{
		Name: "root",
		Type: SchemaTypeObject,
		Properties: map[string]*Item{
			"current_deployment": {
				Name:     "current_deployment",
				Type:     SchemaTypeObject,
				Computed: true,
				Properties: map[string]*Item{
					"status": {Name: "status", Type: SchemaTypeString, Computed: true},
				},
			},
			"application_versions": {
				Name:     "application_versions",
				Type:     SchemaTypeArrayOrdered,
				Computed: true,
				Items: &Item{
					Name:     "application_versions",
					Type:     SchemaTypeObject,
					Computed: true,
					Properties: map[string]*Item{
						"version": {Name: "version", Type: SchemaTypeInteger, Computed: true},
					},
				},
			},
		},
	}
	setParents(root, nil)

	code, err := genSchemaInternal(&Definition{Resource: &SchemaMeta{}}, resourceType, root)
	require.NoError(t, err)
	got := strings.Join(strings.Fields(renderCode(t, code)), " ")
	require.Contains(t, got, `"current_deployment": &adapter.Schema{ Computed: true, IsAttribute: true,`)
	require.Contains(t, got, `"application_versions": &adapter.Schema{ Computed: true, IsAttribute: true,`)
}

// Only the resource's user config block is marked, not its nested blocks or the data source's block.
func TestGenSchemaInternalMarksUserConfig(t *testing.T) {
	userConfig := &Item{
		Name:           "datadog_user_config",
		Type:           SchemaTypeObject,
		UserConfigType: "datadog",
		Computed:       true,
		Properties: map[string]*Item{
			"opensearch": {
				Name:       "opensearch",
				Type:       SchemaTypeObject,
				Computed:   true,
				Properties: map[string]*Item{"enabled": {Name: "enabled", Type: SchemaTypeBoolean, Computed: true}},
			},
		},
	}
	root := &Item{Name: "root", Type: SchemaTypeObject, Properties: map[string]*Item{"datadog_user_config": userConfig}}
	setParents(root, nil)

	for _, entity := range []entityType{resourceType, datasourceType} {
		code, err := genSchemaInternal(&Definition{Resource: &SchemaMeta{}, Datasource: &SchemaMeta{}}, entity, root)
		require.NoError(t, err)
		got := renderCode(t, code)
		if entity == resourceType {
			require.Equal(t, 1, strings.Count(got, "IsUserConfig: true"))
		} else {
			require.NotContains(t, got, "IsUserConfig")
		}
	}
}

// A configurable object stays a block, and keeps the singleton validator.
func TestGenSchemaAddsSingletonValidatorForOptionalObjectBlocks(t *testing.T) {
	got := genResourceSchemaCode(t, rootWithObject(&Item{
		Name:     "config",
		Type:     SchemaTypeObject,
		Optional: true,
		Computed: true,
		Properties: map[string]*Item{
			"retention_ms": {Name: "retention_ms", Type: SchemaTypeString, Optional: true, Computed: true},
		},
	}))

	require.Contains(t, got, `"config": schema.ListNestedBlock{`)
	require.Contains(t, got, `[]validator.List{listvalidator.SizeAtMost(1)}`)
}

// One settable field is enough to keep the block syntax.
func TestGenSchemaKeepsObjectWithOptionalFieldAsBlock(t *testing.T) {
	got := genResourceSchemaCode(t, rootWithObject(&Item{
		Name:     "config",
		Type:     SchemaTypeObject,
		Computed: true,
		Properties: map[string]*Item{
			"retention_ms": {Name: "retention_ms", Type: SchemaTypeString, Optional: true},
		},
	}))

	require.Contains(t, got, `"config": schema.ListNestedBlock{`)
}

func TestGenSchemaAddsCIDRValidatorForCIDRFormat(t *testing.T) {
	root := &Item{
		Name: "root",
		Type: SchemaTypeObject,
		Properties: map[string]*Item{
			"network_cidr": {
				Name:      "network_cidr",
				Type:      SchemaTypeString,
				Required:  true,
				Format:    StringFormatCIDR,
				MaxLength: 18,
			},
		},
	}
	root.Properties["network_cidr"].Parent = root

	code, err := genSchema(&Definition{Resource: &SchemaMeta{}}, resourceType, root)
	require.NoError(t, err)

	got := renderCode(t, code)
	require.Contains(t, got, `"network_cidr": schema.StringAttribute{`)
	require.Contains(t, got, `[]validator.String{validators.CIDR()}`)
	require.NotContains(t, got, "Length")
	require.NotContains(t, got, "length")
}

func TestGenSchemaIPv4FromOpenAPI(t *testing.T) {
	// Only the request declares the format. The read response must not erase it,
	// and no field override in the resource definition should be necessary.
	var doc OpenAPIDoc
	require.NoError(t, json.Unmarshal([]byte(`{
		"paths": {
			"/project/{project}/connection": {
				"post": {
					"operationId": "ConnectionCreate",
					"parameters": [{"$ref":"#/components/parameters/project"}],
					"requestBody": {"content":{"application/json":{"schema":{"$ref":"#/components/schemas/Request"}}}}
				},
				"get": {
					"operationId": "ConnectionGet",
					"parameters": [{"$ref":"#/components/parameters/project"}],
					"responses": {"200":{"content":{"application/json":{"schema":{"$ref":"#/components/schemas/Response"}}}}}
				}
			}
		},
		"components": {
			"parameters": {"project":{"in":"path","name":"project","required":true,"schema":{"type":"string"}}},
			"schemas": {
				"Request": {
					"type":"object", "required":["user_ip_address"],
					"properties":{"user_ip_address":{"type":"string","format":"ipv4","maxLength":15,"description":"Endpoint address."}}
				},
				"Response": {
					"type":"object",
					"properties":{"user_ip_address":{"type":"string","description":"Endpoint address."}}
				}
			}
		}
	}`), &doc))
	def := &Definition{
		Resource:       &SchemaMeta{},
		IdentityFields: []string{"project"},
		Operations: Operations{
			{ID: "ConnectionCreate", Type: OperationCreate},
			{ID: "ConnectionGet", Type: OperationRead},
		},
	}
	root, err := createRootItem(&Scope{OpenAPI: &doc, Definition: def, CurrentMeta: def.Resource})
	require.NoError(t, err)
	require.Equal(t, StringFormatIPv4, root.Properties["user_ip_address"].Format)
	got := genResourceSchemaCode(t, root)
	require.Contains(t, got, `[]validator.String{validators.IPv4()}`)
	require.Contains(t, got, "Endpoint address. Must be a valid IPv4 address.")
	require.NotContains(t, got, "LengthAtMost")
	value, err := exampleScalarItem(def, root.Properties["user_ip_address"])
	require.NoError(t, err)
	require.Equal(t, "10.20.30.40", value.AsString())
}

func TestItemFormat(t *testing.T) {
	t.Parallel()

	list := &Item{Name: "peer_cidrs", Type: SchemaTypeArray}
	list.Items = &Item{Name: "peer_cidrs", Type: SchemaTypeString, Parent: list}

	cases := []struct {
		item *Item
		want StringFormat
	}{
		{item: &Item{Name: "reserved_cidr", Type: SchemaTypeString}, want: StringFormatCIDR},
		{item: &Item{Name: "reserved_cidr", Type: SchemaTypeString, MaxLength: 18}, want: StringFormatIPv4CIDR},
		{item: &Item{Name: "reserved_cidr", Type: SchemaTypeString, MaxLength: 43}, want: StringFormatCIDR},
		{item: &Item{Name: "network_cidr", Type: SchemaTypeString, Format: StringFormatUUID}, want: StringFormatUUID},
		{item: &Item{Name: "address", Type: SchemaTypeString, Format: StringFormatIPv4}, want: StringFormatIPv4},
		{item: &Item{Name: "user_ip_address", Type: SchemaTypeString, MaxLength: 15}, want: ""},
		{item: &Item{Name: "network", Type: SchemaTypeString}, want: ""},
		{item: &Item{Name: "uuid", Type: SchemaTypeString}, want: StringFormatUUID},
		{item: &Item{Name: "peering_connection_id", Type: SchemaTypeString, MinLength: 36, MaxLength: 36}, want: ""},
		{item: &Item{Name: "uuid", Type: SchemaTypeInteger}, want: ""},
		{item: list, want: ""},
		{item: list.Items, want: StringFormatCIDR},
	}
	for _, c := range cases {
		require.Equal(t, c.want, itemFormat(c.item), c.item.Name)
	}
}

func TestGenSchemaAddsCIDRHintWithoutIPExample(t *testing.T) {
	root := &Item{
		Name: "root",
		Type: SchemaTypeObject,
		Properties: map[string]*Item{
			"reserved_cidr": {
				Name:        "reserved_cidr",
				Type:        SchemaTypeString,
				Description: "Reserved range.",
				Required:    true,
			},
			"network_cidr": {
				Name:        "network_cidr",
				Type:        SchemaTypeString,
				Description: "Network range. For example, `192.168.0.0/24`.",
				Required:    true,
				Format:      StringFormatCIDR,
			},
		},
	}
	for _, p := range root.Properties {
		p.Parent = root
	}

	code, err := genSchema(&Definition{Resource: &SchemaMeta{}}, resourceType, root)
	require.NoError(t, err)

	got := renderCode(t, code)
	require.Contains(t, got, "\"Reserved range. Must be a valid CIDR block, for example `10.0.0.0/24`.\"")
	require.Equal(t, 2, strings.Count(got, "validators.CIDR()"))
	require.Contains(t, got, "\"Network range. For example, `192.168.0.0/24`.\"")
}

func TestGenSchemaAddsUUIDValidatorForUUIDFormat(t *testing.T) {
	root := &Item{
		Name: "root",
		Type: SchemaTypeObject,
		Properties: map[string]*Item{
			"tenant_id": {
				Name:        "tenant_id",
				Type:        SchemaTypeString,
				Description: "Tenant ID.",
				Required:    true,
				Format:      StringFormatUUID,
				MinLength:   36,
				MaxLength:   1024,
			},
			"app_id": {
				Name:        "app_id",
				Type:        SchemaTypeString,
				Description: "App ID in UUID4 form.",
				Required:    true,
				Format:      StringFormatUUID,
			},
		},
	}
	for _, p := range root.Properties {
		p.Parent = root
	}

	code, err := genSchema(&Definition{Resource: &SchemaMeta{}}, resourceType, root)
	require.NoError(t, err)

	got := renderCode(t, code)
	require.Contains(t, got, `"tenant_id": schema.StringAttribute{`)
	require.Contains(t, got, `[]validator.String{validators.UUID()}`)
	require.NotContains(t, got, "Length")
	require.NotContains(t, got, "length")
	require.Contains(t, got, `"Tenant ID. Must be a valid UUID."`)
	require.Contains(t, got, `"App ID in UUID4 form."`)
}
