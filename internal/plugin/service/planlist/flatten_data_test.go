package planlist

import (
	"testing"

	"github.com/aiven/go-client-codegen/handler/project"
	"github.com/stretchr/testify/assert"
)

func TestFlattenDataTransformation(t *testing.T) {
	t.Parallel()

	apiResponse := &planList{
		Plans: []project.ServicePlanOut{
			{
				ServicePlan: "business-4",
				ServiceType: "kafka",
				Regions: map[string]any{
					"aws-us-east-1": map[string]any{
						"price": "0.50",
					},
					"aws-eu-west-1": map[string]any{
						"price": "0.55",
					},
					"google-us-central1": map[string]any{
						"price": "0.52",
					},
				},
			},
		},
	}

	state := &tfModel{}
	diags := flattenData(t.Context(), state, apiResponse, regionsToArray)

	assert.False(t, diags.HasError(), "Expected no errors but got: %v", diags)
	assert.False(t, state.ServicePlans.IsNull())
}
