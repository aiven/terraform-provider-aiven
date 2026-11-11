package serviceintegration

import (
	"context"
	"fmt"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func TestCreateViewEndpointIDs(t *testing.T) {
	ctx := context.Background()
	client := avngen.NewMockClient(t)
	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		ResourceOptions.IdentityFields,
		adapter.WithTestPlan(map[string]any{
			"project":                 "my-project",
			"integration_type":        "autoscaler",
			"source_service_name":     "my-pg",
			"destination_endpoint_id": "my-project/endpoint-id",
		}),
	)
	require.NoError(t, err)

	client.EXPECT().
		ServiceIntegrationCreate(ctx, "my-project", mock.MatchedBy(func(req *service.ServiceIntegrationCreateIn) bool {
			return lo.FromPtr(req.DestEndpointId) == "endpoint-id" &&
				lo.FromPtr(req.SourceService) == "my-pg" &&
				req.DestService == nil &&
				req.UserConfig != nil && len(*req.UserConfig) == 0
		})).
		Return(&service.ServiceIntegrationCreateOut{
			DestEndpointId:       lo.ToPtr("endpoint-id"),
			DestProject:          "my-project",
			DestService:          lo.ToPtr("ignored"),
			IntegrationType:      service.IntegrationTypeAutoscaler,
			ServiceIntegrationId: "integration-id",
			SourceProject:        "my-project",
			SourceService:        "my-pg",
		}, nil).
		Once()

	require.NoError(t, createView(ctx, client, d))
	assert.Equal(t, "my-project/integration-id", d.ID())
	assert.Equal(t, "my-project/endpoint-id", d.Get("destination_endpoint_id"))
	assert.Equal(t, "my-pg", d.Get("source_service_name"))
	_, ok := d.GetOk("destination_service_name")
	assert.False(t, ok, "the service name of an endpoint destination is not set")
}

func TestCreateViewUserConfig(t *testing.T) {
	ctx := context.Background()
	client := avngen.NewMockClient(t)
	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		ResourceOptions.IdentityFields,
		adapter.WithTestPlan(map[string]any{
			"project":                  "my-project",
			"integration_type":         "logs",
			"source_service_name":      "my-pg",
			"destination_service_name": "my-os",
			"logs_user_config": []any{map[string]any{
				"elasticsearch_index_days_max": 5,
			}},
		}),
	)
	require.NoError(t, err)

	client.EXPECT().
		ServiceIntegrationCreate(ctx, "my-project", mock.MatchedBy(func(req *service.ServiceIntegrationCreateIn) bool {
			return req.UserConfig != nil && fmt.Sprint((*req.UserConfig)["elasticsearch_index_days_max"]) == "5"
		})).
		Return(&service.ServiceIntegrationCreateOut{
			DestProject:          "my-project",
			DestService:          lo.ToPtr("my-os"),
			IntegrationType:      service.IntegrationTypeLogs,
			ServiceIntegrationId: "integration-id",
			SourceProject:        "my-project",
			SourceService:        "my-pg",
			UserConfig:           map[string]any{"elasticsearch_index_days_max": 5},
		}, nil).
		Once()

	require.NoError(t, createView(ctx, client, d))
	assert.Equal(t, "my-os", d.Get("destination_service_name"))
	assert.EqualValues(t, 5, d.Get("logs_user_config.0.elasticsearch_index_days_max"))
}

func TestCreateViewEmptyServiceName(t *testing.T) {
	d, err := adapter.NewResourceData(resourceSchemaInternal(), ResourceOptions.IdentityFields,
		adapter.WithTestPlan(map[string]any{
			"project": "my-project", "integration_type": "autoscaler",
			"source_service_name": "my-pg", "destination_endpoint_id": "my-project/endpoint-id",
			"destination_service_name": "",
		}))
	require.NoError(t, err)
	client := avngen.NewMockClient(t)
	client.EXPECT().ServiceIntegrationCreate(mock.Anything, "my-project", mock.Anything).
		RunAndReturn(func(_ context.Context, _ string, req *service.ServiceIntegrationCreateIn) (*service.ServiceIntegrationCreateOut, error) {
			assert.Nil(t, req.DestService, "an unused empty service name is omitted from API requests")
			return &service.ServiceIntegrationCreateOut{
				ServiceIntegrationId: "integration-id", IntegrationType: service.IntegrationTypeAutoscaler,
				SourceService: "my-pg", DestEndpointId: lo.ToPtr("endpoint-id"),
			}, nil
		}).Once()
	require.NoError(t, createView(context.Background(), client, d))
	v, ok := d.GetOk("destination_service_name")
	assert.True(t, ok, "the state must keep the explicitly configured empty value")
	assert.Empty(t, v)
}

func TestUpdateUserConfigWithEmptyServiceName(t *testing.T) {
	state := map[string]any{
		"id": "my-project/integration-id", "project": "my-project", "integration_id": "integration-id",
		"integration_type": "datadog", "datadog_user_config": []any{map[string]any{"exclude_topics": []any{"topic"}}},
	}
	plan := map[string]any{
		"id": "my-project/integration-id", "project": "my-project", "integration_id": "integration-id",
		"integration_type": "datadog", "datadog_user_config": []any{map[string]any{"exclude_topics": []any{}}},
		"destination_service_name": "",
	}
	d, err := adapter.NewResourceData(resourceSchemaInternal(), ResourceOptions.IdentityFields, adapter.WithTestState(state), adapter.WithTestPlan(plan))
	require.NoError(t, err)
	client := avngen.NewMockClient(t)
	client.EXPECT().ServiceIntegrationUpdate(mock.Anything, "my-project", "integration-id", mock.Anything).
		RunAndReturn(func(_ context.Context, _, _ string, req *service.ServiceIntegrationUpdateIn) (*service.ServiceIntegrationUpdateOut, error) {
			assert.Equal(t, map[string]any{"exclude_topics": []any{}}, req.UserConfig)
			return &service.ServiceIntegrationUpdateOut{
				IntegrationType: "datadog", UserConfig: req.UserConfig, ServiceIntegrationId: "integration-id",
				SourceProject: "my-project", DestProject: "my-project",
			}, nil
		}).Once()
	require.NoError(t, ResourceOptions.Update(context.Background(), client, d))
}

func TestUpdateEmptyServiceNameWithUnchangedUserConfig(t *testing.T) {
	values := func() map[string]any {
		return map[string]any{
			"id": "my-project/integration-id", "project": "my-project", "integration_id": "integration-id",
			"integration_type": "datadog", "datadog_user_config": []any{map[string]any{"exclude_topics": []any{"topic"}}},
		}
	}
	state := values()
	plan := values()
	plan["destination_service_name"] = ""
	d, err := adapter.NewResourceData(resourceSchemaInternal(), ResourceOptions.IdentityFields, adapter.WithTestState(state), adapter.WithTestPlan(plan))
	require.NoError(t, err)
	client := avngen.NewMockClient(t)
	err = ResourceOptions.Update(context.Background(), client, d)
	require.ErrorIs(t, err, adapter.ErrNoChanges)
}

func TestServiceNamePlanModifiers(t *testing.T) {
	ctx := context.Background()
	attr := ResourceOptions.Schema(ctx).Attributes["source_service_name"].(schema.StringAttribute)
	for _, tc := range []struct {
		name                string
		config, plan, state types.String
		replace             bool
	}{
		{"unconfigured computed name", types.StringNull(), types.StringUnknown(), types.StringValue("source"), false},
		{"configured unknown name", types.StringUnknown(), types.StringUnknown(), types.StringValue("source"), true},
		{"changed name", types.StringValue("other"), types.StringValue("other"), types.StringValue("source"), true},
		{"null to empty", types.StringValue(""), types.StringValue(""), types.StringNull(), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := func(v types.String) tftypes.Value {
				value, err := v.ToTerraformValue(ctx)
				require.NoError(t, err)
				return tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"source_service_name": tftypes.String}}, map[string]tftypes.Value{"source_service_name": value})
			}
			req := planmodifier.StringRequest{
				ConfigValue: tc.config, PlanValue: tc.plan, StateValue: tc.state,
				Config: tfsdk.Config{Raw: raw(tc.config)}, Plan: tfsdk.Plan{Raw: raw(tc.plan)}, State: tfsdk.State{Raw: raw(tc.state)},
			}
			var replace bool
			for _, mod := range attr.PlanModifiers {
				resp := planmodifier.StringResponse{PlanValue: req.PlanValue}
				mod.PlanModifyString(ctx, req, &resp)
				require.False(t, resp.Diagnostics.HasError())
				replace = replace || resp.RequiresReplace
				req.PlanValue = resp.PlanValue
			}
			assert.Equal(t, tc.replace, replace)
		})
	}
}

// An update without the user config would reset the settings: the mock fails on any call.
func TestUpdateWithoutUserConfig(t *testing.T) {
	values := func() map[string]any {
		return map[string]any{
			"id":                       "my-project/integration-id",
			"project":                  "my-project",
			"integration_id":           "integration-id",
			"integration_type":         "logs",
			"source_service_name":      "my-pg",
			"destination_service_name": "my-os",
		}
	}
	state := values()
	state["logs_user_config"] = []any{map[string]any{"elasticsearch_index_days_max": 5}}
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

func TestReadViewPreservesUnusedServiceName(t *testing.T) {
	ctx := context.Background()
	client := avngen.NewMockClient(t)
	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		ResourceOptions.IdentityFields,
		adapter.WithTestState(map[string]any{
			"id":                       "my-project/integration-id",
			"project":                  "my-project",
			"integration_id":           "integration-id",
			"integration_type":         "autoscaler",
			"source_service_name":      "my-pg",
			"source_endpoint_id":       "",
			"destination_endpoint_id":  "my-project/endpoint-id",
			"destination_service_name": "",
		}),
	)
	require.NoError(t, err)

	client.EXPECT().
		ServiceIntegrationGet(ctx, "my-project", "integration-id").
		Return(&service.ServiceIntegrationGetOut{
			DestEndpointId:       lo.ToPtr("endpoint-id"),
			DestProject:          "my-project",
			IntegrationType:      "autoscaler",
			ServiceIntegrationId: "integration-id",
			SourceProject:        "my-project",
			SourceService:        "my-pg",
		}, nil).
		Once()

	require.NoError(t, readView(ctx, client, d))
	assert.Equal(t, "my-project/endpoint-id", d.Get("destination_endpoint_id"))
	_, ok := d.GetOk("source_endpoint_id")
	assert.False(t, ok, "an unused endpoint ID is unset")
	v, ok := d.GetOk("destination_service_name")
	assert.True(t, ok, "an unused service name keeps its prior empty representation")
	assert.Empty(t, v)
}

func TestAdoptReadReplica(t *testing.T) {
	ctx := context.Background()
	client := avngen.NewMockClient(t)
	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		ResourceOptions.IdentityFields,
		adapter.WithTestPlan(map[string]any{
			"project":                  "my-project",
			"integration_type":         "read_replica",
			"source_service_name":      "my-pg",
			"destination_service_name": "my-pg-replica",
		}),
	)
	require.NoError(t, err)

	client.EXPECT().
		ServiceIntegrationList(ctx, "my-project", "my-pg").
		Return([]service.ServiceIntegrationOut{
			{
				DestService:          lo.ToPtr("other-replica"),
				IntegrationType:      service.IntegrationTypeReadReplica,
				ServiceIntegrationId: "other-id",
				SourceService:        "my-pg",
			},
			{
				DestService:          lo.ToPtr("my-pg-replica"),
				IntegrationType:      service.IntegrationTypeReadReplica,
				ServiceIntegrationId: "replica-id",
				SourceService:        "my-pg",
			},
		}, nil).
		Once()

	create := func(context.Context, avngen.Client, adapter.ResourceData) error {
		t.Fatal("a preexisting read_replica integration is not created")
		return nil
	}

	require.NoError(t, adoptReadReplica(create)(ctx, client, d))
	assert.Equal(t, "replica-id", d.Get("integration_id"))
}
