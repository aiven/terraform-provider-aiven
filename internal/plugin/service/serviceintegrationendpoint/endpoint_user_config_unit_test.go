package serviceintegrationendpoint

import (
	"testing"

	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func TestDatadogUserConfigStringRoundTrip(t *testing.T) {
	for _, prefix := range []string{"", "custom"} {
		for _, phase := range []string{"plan", "state"} {
			t.Run(phase+"/prefix="+prefix, func(t *testing.T) {
				values := map[string]any{
					"id":            "my-project/endpoint-id",
					"project":       "my-project",
					"endpoint_id":   "endpoint-id",
					"endpoint_name": "datadog",
					"endpoint_type": "datadog",
					"datadog_user_config": []any{map[string]any{
						"datadog_api_key":   "11111111222233334444555555555555",
						"extra_tags_prefix": prefix,
					}},
				}
				d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestPlan(values))
				require.NoError(t, err)

				var request service.ServiceIntegrationEndpointCreateIn
				require.NoError(t, d.Expand(&request, adapter.ExpandUserConfig("endpoint_type")))
				require.Equal(t, prefix, request.UserConfig["extra_tags_prefix"])
				if phase == "state" {
					d, err = adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestState(values))
					require.NoError(t, err)
				}

				// Exercise the full Flatten path, including normalization after FlattenUserConfig.
				require.NoError(t, d.Flatten(&service.ServiceIntegrationEndpointGetOut{
					EndpointId:   "endpoint-id",
					EndpointName: request.EndpointName,
					EndpointType: request.EndpointType,
					UserConfig:   request.UserConfig,
				}, adapter.FlattenUserConfig("endpoint_type")))

				got, ok := d.GetOk("datadog_user_config.0.extra_tags_prefix")
				require.True(t, ok, "a configured empty string must not become null")
				assert.Equal(t, prefix, got)

				// Use the flattened values as the next update's plan, including after a Read.
				nextPlan, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithTestPlan(values))
				require.NoError(t, err)
				var nextRequest service.ServiceIntegrationEndpointUpdateIn
				require.NoError(t, nextPlan.Expand(&nextRequest, adapter.ExpandUserConfig("endpoint_type")))
				assert.Equal(t, request.UserConfig, nextRequest.UserConfig)
			})
		}
	}
}
