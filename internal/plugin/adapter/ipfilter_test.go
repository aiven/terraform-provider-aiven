package adapter

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ipFilterTestBlock struct {
	IPFilter       types.Set `tfsdk:"ip_filter"`
	IPFilterString types.Set `tfsdk:"ip_filter_string"`
}

type ipFilterTestModel struct {
	Block []ipFilterTestBlock `tfsdk:"block"`
}

func TestIPFilterPlanModifier(t *testing.T) {
	t.Parallel()

	null := types.SetNull(types.StringType)
	set := func(v ...string) types.Set {
		items := make([]attr.Value, len(v))
		for i, s := range v {
			items[i] = types.StringValue(s)
		}
		return types.SetValueMust(types.StringType, items)
	}

	cases := []struct {
		name                   string
		config, sibling, state types.Set
		want                   types.Set
	}{
		{name: "removed list is cleared", config: null, sibling: null, state: set("10.0.0.0/8"), want: set()},
		{name: "removed default list is kept", config: null, sibling: null, state: set("::/0", "0.0.0.0/0"), want: set("::/0", "0.0.0.0/0")},
		{name: "removed IPv4 default list is kept", config: null, sibling: null, state: set("0.0.0.0/0"), want: set("0.0.0.0/0")},
		{name: "removed empty list is kept", config: null, sibling: null, state: set(), want: set()},
		{name: "configured list is kept", config: set("10.0.0.0/8"), sibling: null, state: set("10.0.0.0/8"), want: set("10.0.0.0/8")},
		{name: "no state", config: null, sibling: null, state: null, want: null},
		{name: "replaced with another option", config: null, sibling: set("10.0.0.0/8"), state: set("10.0.0.0/8"), want: set("10.0.0.0/8")},
	}

	sch := schema.Schema{
		Blocks: map[string]schema.Block{
			"block": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"ip_filter":        schema.SetAttribute{ElementType: types.StringType, Optional: true, Computed: true},
						"ip_filter_string": schema.SetAttribute{ElementType: types.StringType, Optional: true, Computed: true},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			plan := tfsdk.Plan{Schema: sch, Raw: tftypes.NewValue(sch.Type().TerraformType(ctx), nil)}
			model := ipFilterTestModel{Block: []ipFilterTestBlock{{IPFilter: tc.config, IPFilterString: tc.sibling}}}
			require.False(t, plan.Set(ctx, &model).HasError())

			req := planmodifier.SetRequest{
				Path:        path.Root("block").AtListIndex(0).AtName("ip_filter"),
				Config:      tfsdk.Config{Schema: sch, Raw: plan.Raw},
				ConfigValue: tc.config,
				StateValue:  tc.state,
				PlanValue:   tc.state,
			}
			rsp := &planmodifier.SetResponse{PlanValue: req.PlanValue}
			IPFilterPlanModifier().PlanModifySet(ctx, req, rsp)
			require.False(t, rsp.Diagnostics.HasError(), rsp.Diagnostics)
			assert.True(t, tc.want.Equal(rsp.PlanValue), "got %s", rsp.PlanValue)
		})
	}
}
