package planlist

import (
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	avivenproject "github.com/aiven/go-client-codegen/handler/project"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadPlanList(t *testing.T) {
	t.Parallel()

	var (
		projectName = "test-project"
		serviceType = "kafka"

		client = avngen.NewMockClient(t)
	)

	state := &tfModel{
		Project:     types.StringValue(projectName),
		ServiceType: types.StringValue(serviceType),
	}

	plansOut := []avivenproject.ServicePlanOut{
		{
			ServicePlan: "business-4",
			ServiceType: serviceType,
			Regions: map[string]any{
				"aws-us-east-1": map[string]any{
					"price": "1.50",
				},
				"aws-eu-west-1": map[string]any{
					"price": "1.55",
				},
				"google-us-central1": map[string]any{
					"price": "1.52",
				},
			},
		},
		{
			ServicePlan: "startup-4",
			ServiceType: serviceType,
			Regions: map[string]any{
				"aws-us-east-1": map[string]any{
					"price": "0.50",
				},
			},
		},
	}

	client.EXPECT().
		ProjectServicePlanList(t.Context(), projectName, serviceType).
		Return(plansOut, nil).
		Once()

	diags := readPlanList(t.Context(), client, state)

	require.False(t, diags.HasError(), "expected no errors but got: %v", diags)

	assert.Equal(t, projectName, state.Project.ValueString())
	assert.Equal(t, serviceType, state.ServiceType.ValueString())

	assert.False(t, state.ServicePlans.IsNull(), "service_plans should not be null")
}
