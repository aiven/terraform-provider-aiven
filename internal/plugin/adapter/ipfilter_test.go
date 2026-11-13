package adapter

import (
	"os"
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
	unknown := types.SetUnknown(types.StringType)
	set := func(v ...string) types.Set {
		items := make([]attr.Value, len(v))
		for i, s := range v {
			items[i] = types.StringValue(s)
		}
		return types.SetValueMust(types.StringType, items)
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

	// An update plans the unset options unknown, as they are computed
	cases := []struct {
		name                         string
		create                       bool
		config, sibling, state, want types.Set
	}{
		{name: "removed list is cleared", config: null, sibling: null, state: set("10.0.0.0/8"), want: set()},
		{name: "removed default list is kept", config: null, sibling: null, state: set("::/0", "0.0.0.0/0"), want: set("::/0", "0.0.0.0/0")},
		{name: "removed IPv4 default list is kept", config: null, sibling: null, state: set("0.0.0.0/0"), want: set("0.0.0.0/0")},
		{name: "removed empty list is kept", config: null, sibling: null, state: set(), want: set()},
		{name: "configured list is kept", config: set("10.0.0.0/8"), sibling: null, state: set("10.0.0.0/8"), want: set("10.0.0.0/8")},
		{name: "no state", config: null, sibling: null, state: null, want: null},
		{name: "create", create: true, config: null, sibling: null, state: null, want: unknown},
		{name: "replaced with another option", config: null, sibling: set("192.168.0.0/16"), state: set("10.0.0.0/8"), want: unknown},
		{name: "another option has the state value", config: null, sibling: set("10.0.0.0/8"), state: set("10.0.0.0/8"), want: set("10.0.0.0/8")},
		{name: "another option is set", config: null, sibling: set("10.0.0.0/8"), state: null, want: null},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			raw := func(ipFilter types.Set) tftypes.Value {
				plan := tfsdk.Plan{Schema: sch, Raw: tftypes.NewValue(sch.Type().TerraformType(ctx), nil)}
				model := ipFilterTestModel{Block: []ipFilterTestBlock{{IPFilter: ipFilter, IPFilterString: tc.sibling}}}
				require.False(t, plan.Set(ctx, &model).HasError())
				return plan.Raw
			}

			state := raw(tc.state)
			if tc.create {
				state = tftypes.NewValue(sch.Type().TerraformType(ctx), nil)
			}

			plan := unknown
			if !tc.config.IsNull() {
				plan = tc.config
			}

			req := planmodifier.SetRequest{
				Path:        path.Root("block").AtListIndex(0).AtName("ip_filter"),
				Config:      tfsdk.Config{Schema: sch, Raw: raw(tc.config)},
				State:       tfsdk.State{Schema: sch, Raw: state},
				ConfigValue: tc.config,
				StateValue:  tc.state,
				PlanValue:   plan,
			}
			rsp := &planmodifier.SetResponse{PlanValue: req.PlanValue}
			IPFilterPlanModifier().PlanModifySet(ctx, req, rsp)
			require.False(t, rsp.Diagnostics.HasError(), rsp.Diagnostics)
			assert.True(t, tc.want.Equal(rsp.PlanValue), "got %s", rsp.PlanValue)
		})
	}
}

func TestSameNetworks(t *testing.T) {
	t.Parallel()

	cidrs := func(v ...string) types.Set {
		items := make([]attr.Value, len(v))
		for i, s := range v {
			items[i] = types.StringValue(s)
		}
		return types.SetValueMust(types.StringType, items)
	}
	objType := map[string]attr.Type{"network": types.StringType, "description": types.StringType}
	objects := func(v ...string) types.Set {
		items := make([]attr.Value, len(v))
		for i, s := range v {
			items[i] = types.ObjectValueMust(objType, map[string]attr.Value{
				"network":     types.StringValue(s),
				"description": types.StringNull(),
			})
		}
		return types.SetValueMust(types.ObjectType{AttrTypes: objType}, items)
	}

	cases := []struct {
		name   string
		option attr.Value
		state  types.Set
		want   bool
	}{
		{name: "strings", option: cidrs("10.0.0.0/8", "::/0"), state: cidrs("::/0", "10.0.0.0/8"), want: true},
		{name: "objects", option: objects("10.0.0.0/8"), state: cidrs("10.0.0.0/8"), want: true},
		{name: "other networks", option: objects("192.168.0.0/16"), state: cidrs("10.0.0.0/8"), want: false},
		{name: "fewer networks", option: cidrs("10.0.0.0/8"), state: cidrs("10.0.0.0/8", "::/0"), want: false},
		{name: "unknown", option: types.SetUnknown(types.StringType), state: cidrs(), want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, sameNetworks(tc.option, tc.state))
		})
	}
}

func TestIsDefaultIPFilter(t *testing.T) {
	t.Parallel()

	cases := []struct {
		cidrs []string
		want  bool
	}{
		{cidrs: []string{"0.0.0.0/0"}, want: true},
		{cidrs: []string{"::/0", "0.0.0.0/0"}, want: true},
		{cidrs: []string{"10.0.0.0/8"}, want: false},
		{cidrs: []string{}, want: false},
		{cidrs: nil, want: false},
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, IsDefaultIPFilter(tc.cidrs))
		})
	}
}

func TestCheckIPFilterPurge(t *testing.T) {
	cases := []struct {
		name    string
		config  map[string]any
		allow   bool
		wantErr bool
	}{
		{name: "no ip_filter", config: map[string]any{}},
		{name: "networks", config: map[string]any{"ip_filter": []any{"10.0.0.0/8"}}},
		{name: "empty list", config: map[string]any{"ip_filter": []any{}}, wantErr: true},
		{name: "empty list allowed", config: map[string]any{"ip_filter": []any{}}, allow: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(AllowIPFilterPurge, "")
			if !tc.allow {
				require.NoError(t, os.Unsetenv(AllowIPFilterPurge))
			}
			err := checkIPFilterPurge(tc.config)
			if tc.wantErr {
				require.ErrorContains(t, err, AllowIPFilterPurge)
				return
			}
			require.NoError(t, err)
		})
	}
}
