package serviceintegration

import (
	"context"
	"fmt"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
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
		idFields(),
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
				req.UserConfig == nil
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
		idFields(),
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

func TestUpdateViewSendsEmptyUserConfig(t *testing.T) {
	ctx := context.Background()
	client := avngen.NewMockClient(t)
	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		idFields(),
		adapter.WithTestPlan(map[string]any{
			"id":                       "my-project/integration-id",
			"project":                  "my-project",
			"integration_id":           "integration-id",
			"integration_type":         "logs",
			"source_service_name":      "my-pg",
			"destination_service_name": "my-os",
		}),
	)
	require.NoError(t, err)

	client.EXPECT().
		ServiceIntegrationUpdate(ctx, "my-project", "integration-id", mock.MatchedBy(func(req *service.ServiceIntegrationUpdateIn) bool {
			return req.UserConfig != nil && len(req.UserConfig) == 0
		})).
		Return(&service.ServiceIntegrationUpdateOut{
			DestProject:          "my-project",
			DestService:          lo.ToPtr("my-os"),
			IntegrationType:      "logs",
			ServiceIntegrationId: "integration-id",
			SourceProject:        "my-project",
			SourceService:        "my-pg",
		}, nil).
		Once()

	require.NoError(t, updateView(ctx, client, d))
}

func TestReadViewClearsEmptyStrings(t *testing.T) {
	ctx := context.Background()
	client := avngen.NewMockClient(t)
	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		idFields(),
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
	for _, k := range []string{"destination_service_name", "source_endpoint_id"} {
		_, ok := d.GetOk(k)
		assert.False(t, ok, "%s is not empty", k)
	}
}

func TestAdoptReadReplica(t *testing.T) {
	ctx := context.Background()
	client := avngen.NewMockClient(t)
	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		idFields(),
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
