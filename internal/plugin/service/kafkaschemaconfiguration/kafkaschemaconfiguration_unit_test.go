package kafkaschemaconfiguration

import (
	"context"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkaschemaregistry"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func TestCreateView_withLevel(t *testing.T) {
	t.Parallel()

	const (
		project     = "test-project-ksc-create-level"
		serviceName = "test-service-ksc-create-level"
	)
	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)
	mockClient.EXPECT().
		ServiceGet(ctx, project, serviceName).
		Return(&service.ServiceGetOut{UserConfig: map[string]any{"schema_registry": true}}, nil).
		Once()
	mockClient.EXPECT().
		ServiceSchemaRegistryGlobalConfigPut(
			ctx,
			project,
			serviceName,
			&kafkaschemaregistry.ServiceSchemaRegistryGlobalConfigPutIn{
				Compatibility: kafkaschemaregistry.CompatibilityTypeForward,
			},
		).
		Return(kafkaschemaregistry.CompatibilityTypeForward, nil).
		Once()

	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		ResourceOptions.IdentityFields,
		adapter.WithTestPlan(map[string]any{
			"project":             project,
			"service_name":        serviceName,
			"compatibility_level": "FORWARD",
		}),
	)
	require.NoError(t, err)
	require.NoError(t, createView(ctx, mockClient, d))
	require.Equal(t, project+"/"+serviceName, d.ID())
}

func TestCreateView_withoutLevel(t *testing.T) {
	t.Parallel()

	const (
		project     = "test-project-ksc-create-empty"
		serviceName = "test-service-ksc-create-empty"
	)
	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)
	mockClient.EXPECT().
		ServiceGet(ctx, project, serviceName).
		Return(&service.ServiceGetOut{UserConfig: map[string]any{"schema_registry": true}}, nil).
		Once()

	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		ResourceOptions.IdentityFields,
		adapter.WithTestPlan(map[string]any{
			"project":      project,
			"service_name": serviceName,
		}),
	)
	require.NoError(t, err)
	require.NoError(t, createView(ctx, mockClient, d))
	require.Equal(t, project+"/"+serviceName, d.ID())
}

func TestCreateView_schemaRegistryDisabled(t *testing.T) {
	t.Parallel()

	const (
		project     = "test-project-ksc-create-disabled"
		serviceName = "test-service-ksc-create-disabled"
	)
	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)
	mockClient.EXPECT().
		ServiceGet(ctx, project, serviceName).
		Return(&service.ServiceGetOut{UserConfig: map[string]any{"schema_registry": false}}, nil).
		Once()

	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		ResourceOptions.IdentityFields,
		adapter.WithTestPlan(map[string]any{
			"project":      project,
			"service_name": serviceName,
		}),
	)
	require.NoError(t, err)
	require.ErrorIs(t, createView(ctx, mockClient, d), errRegistryDisabled)
}

func TestReadView_schemaRegistryDisabled(t *testing.T) {
	t.Parallel()

	const (
		project     = "test-project-ksc-read-disabled"
		serviceName = "test-service-ksc-read-disabled"
	)
	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)
	mockClient.EXPECT().
		ServiceGet(ctx, project, serviceName).
		Return(&service.ServiceGetOut{UserConfig: map[string]any{"schema_registry": false}}, nil).
		Once()

	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		ResourceOptions.IdentityFields,
		adapter.WithTestState(map[string]any{
			"id":           project + "/" + serviceName,
			"project":      project,
			"service_name": serviceName,
		}),
	)
	require.NoError(t, err)
	err = readView(ctx, mockClient, d)
	require.Error(t, err)
	require.True(t, adapter.IsNotFound(err))
}

func TestDeleteView_resetsBackward(t *testing.T) {
	t.Parallel()

	const (
		project     = "test-project-ksc-delete"
		serviceName = "test-service-ksc-delete"
	)
	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)
	mockClient.EXPECT().
		ServiceGet(ctx, project, serviceName).
		Return(&service.ServiceGetOut{UserConfig: map[string]any{"schema_registry": true}}, nil).
		Once()
	mockClient.EXPECT().
		ServiceSchemaRegistryGlobalConfigPut(
			ctx,
			project,
			serviceName,
			&kafkaschemaregistry.ServiceSchemaRegistryGlobalConfigPutIn{
				Compatibility: kafkaschemaregistry.CompatibilityTypeBackward,
			},
		).
		Return(kafkaschemaregistry.CompatibilityTypeBackward, nil).
		Once()

	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		ResourceOptions.IdentityFields,
		adapter.WithTestState(map[string]any{
			"id":                  project + "/" + serviceName,
			"project":             project,
			"service_name":        serviceName,
			"compatibility_level": "FORWARD",
		}),
	)
	require.NoError(t, err)
	require.NoError(t, deleteView(ctx, mockClient, d))
}

func TestDeleteView_schemaRegistryDisabled(t *testing.T) {
	t.Parallel()

	const (
		project     = "test-project-ksc-delete-disabled"
		serviceName = "test-service-ksc-delete-disabled"
	)
	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)
	mockClient.EXPECT().
		ServiceGet(ctx, project, serviceName).
		Return(&service.ServiceGetOut{UserConfig: map[string]any{"schema_registry": false}}, nil).
		Once()

	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		ResourceOptions.IdentityFields,
		adapter.WithTestState(map[string]any{
			"id":           project + "/" + serviceName,
			"project":      project,
			"service_name": serviceName,
		}),
	)
	require.NoError(t, err)
	require.NoError(t, deleteView(ctx, mockClient, d))
}
