package vpc_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/awsvpcpeeringconnection"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/azurevpcpeeringconnection"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/gcpvpcpeeringconnection"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/transitgatewayvpcattachment"
)

const validVPCUUID = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"

func TestProjectPeeringVPCSelectorsValidation(t *testing.T) {
	for name, config := range peeringSchemas(t) {
		t.Run(name, func(t *testing.T) {
			for _, field := range []string{"vpc_id", "project_vpc_id"} {
				t.Run(field, func(t *testing.T) {
					prefix := ""
					if field == "vpc_id" {
						prefix = "project/"
					}
					t.Run("lowercase", func(t *testing.T) {
						checkPeeringString(t, config, field, types.StringValue(prefix+validVPCUUID), false)
					})
					t.Run("uppercase", func(t *testing.T) {
						checkPeeringString(t, config, field, types.StringValue(prefix+strings.ToUpper(validVPCUUID)), true)
					})
					t.Run("malformed UUID", func(t *testing.T) {
						checkPeeringString(t, config, field, types.StringValue(prefix+"not-a-uuid"), true)
					})
					t.Run("unknown", func(t *testing.T) {
						checkPeeringString(t, config, field, types.StringUnknown(), false)
					})
				})
			}
		})
	}
}

func TestAzurePeeringFieldsValidation(t *testing.T) {
	for name, config := range peeringSchemas(t) {
		if !strings.HasPrefix(name, "azure") {
			continue
		}
		t.Run(name, func(t *testing.T) {
			for _, field := range []string{"azure_subscription_id", "peer_azure_app_id", "peer_azure_tenant_id"} {
				t.Run(field, func(t *testing.T) {
					checkPeeringString(t, config, field, types.StringValue(validVPCUUID), false)
					checkPeeringString(t, config, field, types.StringValue(""), true)
					checkPeeringString(t, config, field, types.StringValue(strings.ToUpper(validVPCUUID)), true)
					checkPeeringString(t, config, field, types.StringUnknown(), false)
				})
			}
			t.Run("VNet name", func(t *testing.T) {
				for _, value := range []string{"my-vnet", "a0", strings.Repeat("a", 64), "a.b_c-d"} {
					checkPeeringString(t, config, "vnet_name", types.StringValue(value), false)
				}
				for _, value := range []string{"My-Vnet", "a", strings.Repeat("a", 65), "-vnet", "vnet-"} {
					checkPeeringString(t, config, "vnet_name", types.StringValue(value), true)
				}
				checkPeeringString(t, config, "vnet_name", types.StringUnknown(), false)
			})
			t.Run("resource group", func(t *testing.T) {
				for _, value := range []string{"my-group", "a", strings.Repeat("a", 90), "group.(test)_1"} {
					checkPeeringString(t, config, "peer_resource_group", types.StringValue(value), false)
				}
				for _, value := range []string{"My-Group", "", strings.Repeat("a", 91), "group.", "group/name"} {
					checkPeeringString(t, config, "peer_resource_group", types.StringValue(value), true)
				}
				checkPeeringString(t, config, "peer_resource_group", types.StringUnknown(), false)
			})
		})
	}
}

func peeringSchemas(t *testing.T) map[string]tfsdk.Config {
	t.Helper()
	return map[string]tfsdk.Config{
		"aws resource":      {Schema: awsvpcpeeringconnection.ResourceOptions.Schema(t.Context())},
		"aws data source":   {Schema: awsvpcpeeringconnection.DataSourceOptions.Schema(t.Context())},
		"azure resource":    {Schema: azurevpcpeeringconnection.ResourceOptions.Schema(t.Context())},
		"azure data source": {Schema: azurevpcpeeringconnection.DataSourceOptions.Schema(t.Context())},
		"gcp resource":      {Schema: gcpvpcpeeringconnection.ResourceOptions.Schema(t.Context())},
		"gcp data source":   {Schema: gcpvpcpeeringconnection.DataSourceOptions.Schema(t.Context())},
		"tgw resource":      {Schema: transitgatewayvpcattachment.ResourceOptions.Schema(t.Context())},
		"tgw data source":   {Schema: transitgatewayvpcattachment.DataSourceOptions.Schema(t.Context())},
	}
}

func checkPeeringString(t *testing.T, config tfsdk.Config, field string, value types.String, wantError bool) {
	t.Helper()
	typ := config.Schema.Type().TerraformType(t.Context()).(tftypes.Object)
	attributes := make(map[string]tftypes.Value, len(typ.AttributeTypes))
	for name, attributeType := range typ.AttributeTypes {
		attributes[name] = tftypes.NewValue(attributeType, nil)
	}
	if field == "project_vpc_id" {
		attributes["project"] = tftypes.NewValue(tftypes.String, "project")
	}
	var err error
	attributes[field], err = value.ToTerraformValue(t.Context())
	require.NoError(t, err)
	config.Raw = tftypes.NewValue(typ, attributes)
	request := validator.StringRequest{
		Path: path.Root(field), PathExpression: path.MatchRoot(field), Config: config, ConfigValue: value,
	}
	var response validator.StringResponse
	attribute := config.Schema.GetAttributes()[field].(interface{ StringValidators() []validator.String })
	for _, v := range attribute.StringValidators() {
		v.ValidateString(t.Context(), request, &response)
	}
	require.Equal(t, wantError, response.Diagnostics.HasError(), "%s = %s: %v", field, value, response.Diagnostics)
}
