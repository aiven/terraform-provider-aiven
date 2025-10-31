package plan

import (
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	avivenproject "github.com/aiven/go-client-codegen/handler/project"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadPlan(t *testing.T) {
	t.Parallel()

	var (
		projectName = "test-project"
		serviceType = "kafka"
		servicePlan = "business-4"
		cloudName   = "aws-us-east-1"

		client = avngen.NewMockClient(t)
	)

	state := &tfModel{
		Project:     types.StringValue(projectName),
		ServiceType: types.StringValue(serviceType),
		ServicePlan: types.StringValue(servicePlan),
		CloudName:   types.StringValue(cloudName),
	}

	specsOut := &avivenproject.ProjectServicePlanSpecsGetOut{
		BackupConfig: avivenproject.BackupConfigOut{
			Interval:     100,
			MaxCount:     10,
			RecoveryMode: "pitr",
		},
		DiskSpaceCapMb:   lo.ToPtr(1024000),
		DiskSpaceMb:      512000,
		DiskSpaceStepMb:  lo.ToPtr(1024),
		MaxMemoryPercent: lo.ToPtr(90),
		NodeCount:        3,
		ServicePlan:      servicePlan,
		ServiceType:      serviceType,
		ShardCount:       lo.ToPtr(2),
	}

	priceOut := &avivenproject.ProjectServicePlanPriceGetOut{
		BasePriceUsd:            "1.50",
		CloudName:               cloudName,
		ObjectStorageGbPriceUsd: lo.ToPtr("0.025"),
		ServicePlan:             servicePlan,
		ServiceType:             serviceType,
	}

	client.EXPECT().
		ProjectServicePlanSpecsGet(t.Context(), projectName, serviceType, servicePlan).
		Return(specsOut, nil).
		Once()

	client.EXPECT().
		ProjectServicePlanPriceGet(t.Context(), projectName, serviceType, servicePlan, cloudName).
		Return(priceOut, nil).
		Once()

	diags := readPlan(t.Context(), client, state)

	require.False(t, diags.HasError(), "expected no errors but got: %v", diags)

	assert.Equal(t, projectName, state.Project.ValueString())
	assert.Equal(t, serviceType, state.ServiceType.ValueString())
	assert.Equal(t, servicePlan, state.ServicePlan.ValueString())
	assert.Equal(t, cloudName, state.CloudName.ValueString())

	assert.Equal(t, "1.50", state.BasePriceUsd.ValueString())
	assert.Equal(t, "0.025", state.ObjectStorageGbPriceUsd.ValueString())

	assert.Equal(t, int64(3), state.NodeCount.ValueInt64())
	assert.Equal(t, int64(512000), state.DiskSpaceMb.ValueInt64())
	assert.Equal(t, int64(1024000), state.DiskSpaceCapMb.ValueInt64())
	assert.Equal(t, int64(1024), state.DiskSpaceStepMb.ValueInt64())
	assert.Equal(t, int64(90), state.MaxMemoryPercent.ValueInt64())
	assert.Equal(t, int64(2), state.ShardCount.ValueInt64())

	assert.False(t, state.BackupConfig.IsNull(), "backup_config should not be null")
}
