package serviceintegrationendpoint

import (
	"context"
	"net/http"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func newResourceData(t *testing.T) adapter.ResourceData {
	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		ResourceOptions.IdentityFields,
		adapter.WithTestState(map[string]any{
			"id":          "my-project/endpoint-id",
			"project":     "my-project",
			"endpoint_id": "endpoint-id",
		}),
	)
	require.NoError(t, err)
	return d
}

func TestNotFound(t *testing.T) {
	missing := avngen.Error{
		Status:  http.StatusBadRequest,
		Message: "Integration endpoint 'endpoint-id' does not exist",
	}
	invalid := avngen.Error{
		Status:  http.StatusBadRequest,
		Message: "Invalid input",
	}

	cases := []struct {
		name     string
		err      error
		notFound bool
	}{
		{name: "missing endpoint", err: missing, notFound: true},
		{name: "other bad request", err: invalid, notFound: false},
	}

	for _, c := range cases {
		t.Run("read "+c.name, func(t *testing.T) {
			ctx := context.Background()
			client := avngen.NewMockClient(t)
			client.EXPECT().
				ServiceIntegrationEndpointGet(ctx, "my-project", "endpoint-id", [][2]string{{"include_secrets", "true"}}).
				Return(nil, c.err).
				Once()

			err := ResourceOptions.Read(ctx, client, newResourceData(t))
			require.ErrorIs(t, err, c.err)
			assert.Equal(t, c.notFound, adapter.IsNotFound(err))
		})

		t.Run("delete "+c.name, func(t *testing.T) {
			ctx := context.Background()
			client := avngen.NewMockClient(t)
			client.EXPECT().
				ServiceIntegrationEndpointDelete(ctx, "my-project", "endpoint-id").
				Return(c.err).
				Once()

			err := ResourceOptions.Delete(ctx, client, newResourceData(t))
			require.ErrorIs(t, err, c.err)
			assert.Equal(t, c.notFound, adapter.IsNotFound(err))
		})
	}
}

// An empty string is a value: SDKv2 sent and stored it, e.g., a datadog endpoint without the tags prefix.
func TestUserConfigEmptyString(t *testing.T) {
	values := func() map[string]any {
		return map[string]any{
			"id":            "my-project/endpoint-id",
			"project":       "my-project",
			"endpoint_id":   "endpoint-id",
			"endpoint_name": "my-endpoint",
			"endpoint_type": "datadog",
			"datadog_user_config": []any{map[string]any{
				"datadog_api_key":   "11111111222233334444555555555555",
				"extra_tags_prefix": "",
			}},
		}
	}

	create := values()
	delete(create, "id")
	delete(create, "endpoint_id")
	plan, err := adapter.NewResourceData(resourceSchemaInternal(), ResourceOptions.IdentityFields, adapter.WithTestPlan(create))
	require.NoError(t, err)
	req := new(service.ServiceIntegrationEndpointCreateIn)
	require.NoError(t, plan.Expand(req, adapter.ExpandUserConfig("endpoint_type")))
	require.Equal(t, map[string]any{
		"datadog_api_key":   "11111111222233334444555555555555",
		"extra_tags_prefix": "",
	}, req.UserConfig)

	rsp := &service.ServiceIntegrationEndpointGetOut{
		EndpointId:   "endpoint-id",
		EndpointName: "my-endpoint",
		EndpointType: "datadog",
		UserConfig:   req.UserConfig,
	}
	for name, opt := range map[string]func(map[string]any) adapter.ResourceDataOpt{
		"create": adapter.WithTestPlan,
		"read":   adapter.WithTestState,
	} {
		t.Run(name, func(t *testing.T) {
			d, err := adapter.NewResourceData(resourceSchemaInternal(), ResourceOptions.IdentityFields, opt(values()))
			require.NoError(t, err)
			require.NoError(t, d.Flatten(rsp, adapter.FlattenUserConfig("endpoint_type")))
			assert.Equal(t, values()["datadog_user_config"], d.Get("datadog_user_config"))

			// The next plan has the flattened block, and sends the same config
			next := values()
			next["datadog_user_config"] = d.Get("datadog_user_config")
			nextPlan, err := adapter.NewResourceData(
				resourceSchemaInternal(),
				ResourceOptions.IdentityFields,
				adapter.WithTestPlan(next),
				adapter.WithTestState(next),
				adapter.WithTestConfig(values()),
			)
			require.NoError(t, err)
			nextReq := new(service.ServiceIntegrationEndpointUpdateIn)
			require.NoError(t, nextPlan.Expand(nextReq, adapter.ExpandUserConfig("endpoint_type")))
			assert.Equal(t, req.UserConfig, nextReq.UserConfig)
		})
	}
}

// An update without the user config would remove the API key: the mock fails on any call.
func TestUpdateWithoutUserConfig(t *testing.T) {
	values := func() map[string]any {
		return map[string]any{
			"id":            "my-project/endpoint-id",
			"project":       "my-project",
			"endpoint_id":   "endpoint-id",
			"endpoint_name": "my-endpoint",
			"endpoint_type": "datadog",
		}
	}
	state := values()
	state["datadog_user_config"] = []any{map[string]any{"datadog_api_key": "11111111222233334444555555555555"}}
	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		ResourceOptions.IdentityFields,
		adapter.WithTestPlan(values()),
		adapter.WithTestState(state),
	)
	require.NoError(t, err)
	err = ResourceOptions.Update(context.Background(), avngen.NewMockClient(t), d)
	require.ErrorIs(t, err, adapter.ErrNoChanges)
}
