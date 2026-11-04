package pg_test

import (
	"context"
	"maps"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/staticip"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin"
)

const (
	testProject = "my-project"
	testService = "my-pg"
	mockHost    = "my-pg.aivencloud.com"
)

// TestPlanKeepsOutputsKnown guards the plan of a state upgraded from SDKv2.
func TestPlanKeepsOutputsKnown(t *testing.T) {
	ctx := context.Background()
	client := avngen.NewMockClient(t)
	client.EXPECT().
		StaticIPList(mock.Anything, testProject).
		Return([]staticip.StaticIPOut{}, nil)

	server, err := providerserver.NewProtocol6WithError(&plugin.AivenProvider{GenClient: client})()
	require.NoError(t, err)

	schemas, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	requireNoErrorDiagnostics(t, schemas.Diagnostics)

	resourceSchema := schemas.ResourceSchemas["aiven_pg"]
	resourceType := resourceSchema.ValueType()
	providerType := schemas.Provider.ValueType()
	providerConfig := dynamicValue(t, providerType, valueForBlock(schemas.Provider.Block, nil))

	configure, err := server.ConfigureProvider(ctx, &tfprotov6.ConfigureProviderRequest{Config: &providerConfig})
	require.NoError(t, err)
	requireNoErrorDiagnostics(t, configure.Diagnostics)

	uri := "postgres://avnadmin:secret@my-pg.aivencloud.com:12345/defaultdb"
	priorData := map[string]any{
		"cloud_name":             "google-europe-west1",
		"disk_space_used":        "80GiB",
		"id":                     "my-project/my-pg",
		"plan":                   "startup-4",
		"project":                testProject,
		"project_vpc_id":         "",
		"service_host":           mockHost,
		"service_integrations":   []any{},
		"service_name":           testService,
		"service_password":       "secret",
		"service_port":           int64(12345),
		"service_type":           "pg",
		"service_uri":            uri,
		"service_username":       "avnadmin",
		"tag":                    []any{map[string]any{"key": "env", "value": "dev"}},
		"termination_protection": false,
	}

	tests := []struct {
		name           string
		disk           string
		plan           string
		tag            string
		outputsUnknown bool
	}{
		{
			name: "tag update",
			plan: "startup-4",
			tag:  "prod",
		},
		{
			// Older SDKv2 versions stored it for no additional disk
			name: "tag update with 0B disk",
			disk: "0B",
			plan: "startup-4",
			tag:  "prod",
		},
		{
			name:           "plan update",
			plan:           "business-4",
			tag:            "dev",
			outputsUnknown: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priorData := maps.Clone(priorData)
			priorData["additional_disk_space"] = tt.disk
			prior := valueForBlock(resourceSchema.Block, priorData)
			configData := map[string]any{
				"plan":         tt.plan,
				"project":      testProject,
				"service_name": testService,
				"tag":          []any{map[string]any{"key": "env", "value": tt.tag}},
			}
			config := valueForBlock(resourceSchema.Block, configData)
			proposed := proposedState(resourceSchema.Block, configData, priorData)
			configValue := dynamicValue(t, resourceType, config)
			priorValue := dynamicValue(t, resourceType, prior)
			proposedValue := dynamicValue(t, resourceType, proposed)

			resp, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
				TypeName:         "aiven_pg",
				Config:           &configValue,
				PriorState:       &priorValue,
				ProposedNewState: &proposedValue,
			})
			require.NoError(t, err)
			requireNoErrorDiagnostics(t, resp.Diagnostics)

			plannedValue, err := resp.PlannedState.Unmarshal(resourceType)
			require.NoError(t, err)
			var planned map[string]tftypes.Value
			require.NoError(t, plannedValue.As(&planned))

			if tt.outputsUnknown {
				assert.False(t, planned["additional_disk_space"].IsKnown())
				assert.False(t, planned["service_host"].IsKnown())
				return
			}

			assertKnownValue(t, planned["additional_disk_space"], tt.disk)
			assertKnownValue(t, planned["service_host"], mockHost)
			assertKnownValue(t, planned["service_username"], "avnadmin")
			assertKnownValue(t, planned["service_password"], "secret")
		})
	}
}

func dynamicValue(t *testing.T, typ tftypes.Type, value tftypes.Value) tfprotov6.DynamicValue {
	t.Helper()
	result, err := tfprotov6.NewDynamicValue(typ, value)
	require.NoError(t, err)
	return result
}

func valueForBlock(block *tfprotov6.SchemaBlock, data map[string]any) tftypes.Value {
	values := make(map[string]tftypes.Value, len(block.Attributes)+len(block.BlockTypes))
	for _, attribute := range block.Attributes {
		values[attribute.Name] = valueForType(attribute.ValueType(), data[attribute.Name])
	}
	for _, nested := range block.BlockTypes {
		value, ok := data[nested.TypeName]
		if !ok {
			switch nested.Nesting {
			case tfprotov6.SchemaNestedBlockNestingModeList, tfprotov6.SchemaNestedBlockNestingModeSet:
				value = []any{}
			}
		}
		values[nested.TypeName] = valueForType(nested.ValueType(), value)
	}
	return tftypes.NewValue(block.ValueType(), values)
}

func valueForType(typ tftypes.Type, value any) tftypes.Value {
	if value == nil {
		return tftypes.NewValue(typ, nil)
	}

	switch typed := typ.(type) {
	case tftypes.Object:
		input := value.(map[string]any)
		values := make(map[string]tftypes.Value, len(typed.AttributeTypes))
		for name, attributeType := range typed.AttributeTypes {
			values[name] = valueForType(attributeType, input[name])
		}
		return tftypes.NewValue(typ, values)
	case tftypes.List:
		return tftypes.NewValue(typ, collectionValues(typed.ElementType, value))
	case tftypes.Set:
		return tftypes.NewValue(typ, collectionValues(typed.ElementType, value))
	case tftypes.Map:
		input := value.(map[string]any)
		values := make(map[string]tftypes.Value, len(input))
		for key, item := range input {
			values[key] = valueForType(typed.ElementType, item)
		}
		return tftypes.NewValue(typ, values)
	default:
		return tftypes.NewValue(typ, value)
	}
}

func collectionValues(elementType tftypes.Type, value any) []tftypes.Value {
	input := value.([]any)
	values := make([]tftypes.Value, len(input))
	for i, item := range input {
		values[i] = valueForType(elementType, item)
	}
	return values
}

func proposedState(block *tfprotov6.SchemaBlock, config, prior map[string]any) tftypes.Value {
	values := make(map[string]any, len(block.Attributes)+len(block.BlockTypes))
	for _, attribute := range block.Attributes {
		if value, ok := config[attribute.Name]; ok {
			values[attribute.Name] = value
		} else if attribute.Computed {
			values[attribute.Name] = prior[attribute.Name]
		}
	}
	for _, nested := range block.BlockTypes {
		if value, ok := config[nested.TypeName]; ok {
			values[nested.TypeName] = value
		}
	}
	return valueForBlock(block, values)
}

func requireNoErrorDiagnostics(t *testing.T, diagnostics []*tfprotov6.Diagnostic) {
	t.Helper()
	for _, diagnostic := range diagnostics {
		require.NotEqual(t, tfprotov6.DiagnosticSeverityError, diagnostic.Severity, "%s: %s", diagnostic.Summary, diagnostic.Detail)
	}
}

func assertKnownValue(t *testing.T, value tftypes.Value, expected string) {
	t.Helper()
	if !assert.True(t, value.IsKnown()) {
		return
	}
	var actual string
	require.NoError(t, value.As(&actual))
	assert.Equal(t, expected, actual)
}
