package adapter

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// EmptyIsNullPlanModifier plans a string removed from the config as null, or keeps the state when it's null or "":
// both mean unset, e.g., SDKv2 stored "" for an unset optional field.
// todo: drop in v5.0.0, when the state written by SDKv2 is no longer supported.
func EmptyIsNullPlanModifier() planmodifier.String {
	return emptyIsNullPlanModifier{}
}

type emptyIsNullPlanModifier struct{}

func (emptyIsNullPlanModifier) Description(context.Context) string {
	return `Plans a removed value as null, or keeps the state when it's null or "".`
}

func (m emptyIsNullPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (emptyIsNullPlanModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, rsp *planmodifier.StringResponse) {
	if !req.ConfigValue.IsNull() {
		return
	}

	if req.StateValue.IsNull() || req.StateValue.ValueString() == "" {
		rsp.PlanValue = req.StateValue
		return
	}
	rsp.PlanValue = types.StringNull()
}
