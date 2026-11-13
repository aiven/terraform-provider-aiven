package adapter

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestEmptyIsNullPlanModifier(t *testing.T) {
	t.Parallel()

	null, unknown := types.StringNull(), types.StringUnknown()
	str := types.StringValue
	cases := []struct {
		name                string
		config, state, want types.String
	}{
		{name: "create", config: null, state: null, want: null},
		{name: "SDKv2 empty string is kept", config: null, state: str(""), want: str("")},
		{name: "removed value is null", config: null, state: str("my-project/vpc"), want: null},
		{name: "configured empty string", config: str(""), state: null, want: str("")},
		{name: "configured value", config: str("my-project/vpc"), state: str(""), want: str("my-project/vpc")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			plan := unknown
			if !tc.config.IsNull() {
				plan = tc.config
			}
			req := planmodifier.StringRequest{ConfigValue: tc.config, StateValue: tc.state, PlanValue: plan}
			rsp := &planmodifier.StringResponse{PlanValue: plan}
			EmptyIsNullPlanModifier().PlanModifyString(t.Context(), req, rsp)
			assert.True(t, tc.want.Equal(rsp.PlanValue), "got %s", rsp.PlanValue)
		})
	}
}
