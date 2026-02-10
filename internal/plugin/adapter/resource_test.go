package adapter

import (
	"context"
	"fmt"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

type testModel struct {
	ID                    types.String `tfsdk:"id"`
	Project               types.String `tfsdk:"project"`
	ServiceName           types.String `tfsdk:"service_name"`
	DatabaseName          types.String `tfsdk:"database_name"`
	TerminationProtection types.Bool   `tfsdk:"termination_protection"`
	Timeouts              types.Object `tfsdk:"timeouts"`
}

func (m *testModel) SharedModel() *testModel {
	return m
}

func (m *testModel) TimeoutsObject() types.Object {
	return m.Timeouts
}

var testTimeoutAttrTypes = map[string]attr.Type{
	"create":  types.StringType,
	"read":    types.StringType,
	"update":  types.StringType,
	"delete":  types.StringType,
	"default": types.StringType,
}

func testResourceSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"project": schema.StringAttribute{
				Optional: true,
			},
			"service_name": schema.StringAttribute{
				Optional: true,
			},
			"database_name": schema.StringAttribute{
				Optional: true,
			},
			"termination_protection": schema.BoolAttribute{
				Optional: true,
			},
			"timeouts": schema.ObjectAttribute{
				Optional:       true,
				AttributeTypes: testTimeoutAttrTypes,
			},
		},
	}
}

func testTimeoutsObject(t *testing.T, kv ...string) types.Object {
	t.Helper()

	values := map[string]attr.Value{
		"create":  types.StringNull(),
		"read":    types.StringNull(),
		"update":  types.StringNull(),
		"delete":  types.StringNull(),
		"default": types.StringNull(),
	}

	for i := 0; i < len(kv); i += 2 {
		if i+1 < len(kv) {
			values[kv[i]] = types.StringValue(kv[i+1])
		}
	}

	timeoutsObj, diags := types.ObjectValue(testTimeoutAttrTypes, values)
	require.False(t, diags.HasError())
	return timeoutsObj
}

func testNullTerraformValue(ctx context.Context, s schema.Schema) tftypes.Value {
	return tftypes.NewValue(s.Type().TerraformType(ctx), nil)
}

func testPlanFromModel(t *testing.T, ctx context.Context, s schema.Schema, m *testModel) tfsdk.Plan {
	t.Helper()

	plan := tfsdk.Plan{
		Schema: s,
		Raw:    testNullTerraformValue(ctx, s),
	}

	diags := plan.Set(ctx, m)
	require.False(t, diags.HasError())
	return plan
}

func testStateFromModel(t *testing.T, ctx context.Context, s schema.Schema, m *testModel) tfsdk.State {
	t.Helper()

	state := testNullState(ctx, s)

	diags := state.Set(ctx, m)
	require.False(t, diags.HasError())
	return state
}

func testNullState(ctx context.Context, s schema.Schema) tfsdk.State {
	return tfsdk.State{
		Schema: s,
		Raw:    testNullTerraformValue(ctx, s),
	}
}

func testConfigFromPlan(s schema.Schema, plan tfsdk.Plan) tfsdk.Config {
	return tfsdk.Config{
		Schema: s,
		Raw:    plan.Raw,
	}
}

func mockResourceOptions(m *mock.Mock) func(opts *ResourceOptions[*testModel, testModel]) {
	return func(opts *ResourceOptions[*testModel, testModel]) {
		opts.Create = func(ctx context.Context, cl avngen.Client, plan, config *testModel) diag.Diagnostics {
			args := m.MethodCalled("Create", ctx, cl, plan, config)
			return args.Get(0).(diag.Diagnostics)
		}
		opts.Read = func(ctx context.Context, cl avngen.Client, plan *testModel) diag.Diagnostics {
			args := m.MethodCalled("Read", ctx, cl, plan)
			return args.Get(0).(diag.Diagnostics)
		}
		opts.Update = func(ctx context.Context, cl avngen.Client, plan, state, config *testModel) diag.Diagnostics {
			args := m.MethodCalled("Update", ctx, cl, plan, state, config)
			return args.Get(0).(diag.Diagnostics)
		}
		opts.Delete = func(ctx context.Context, cl avngen.Client, state *testModel) diag.Diagnostics {
			args := m.MethodCalled("Delete", ctx, cl, state)
			return args.Get(0).(diag.Diagnostics)
		}
		opts.ConfigValidators = func(ctx context.Context, cl avngen.Client) []resource.ConfigValidator {
			args := m.MethodCalled("ConfigValidators", ctx, cl)
			return args.Get(0).([]resource.ConfigValidator)
		}
		opts.ValidateConfig = func(ctx context.Context, cl avngen.Client, config *testModel) diag.Diagnostics {
			args := m.MethodCalled("ValidateConfig", ctx, cl, config)
			return args.Get(0).(diag.Diagnostics)
		}
		opts.ModifyPlan = func(ctx context.Context, cl avngen.Client, plan, state, config *testModel) diag.Diagnostics {
			args := m.MethodCalled("ModifyPlan", ctx, cl, plan, state, config)
			return args.Get(0).(diag.Diagnostics)
		}
	}
}

func withRefreshState(opts *ResourceOptions[*testModel, testModel]) {
	opts.RefreshState = true
}

func withRemoveMissing(opts *ResourceOptions[*testModel, testModel]) {
	opts.RemoveMissing = true
}

func withBeta(opts *ResourceOptions[*testModel, testModel]) {
	opts.Beta = true
}

func withTerminationProtection(opts *ResourceOptions[*testModel, testModel]) {
	opts.TerminationProtection = true
}

func withPanicModifyPlan(opts *ResourceOptions[*testModel, testModel]) {
	opts.ModifyPlan = func(ctx context.Context, client avngen.Client, plan, state, config *testModel) diag.Diagnostics {
		panic("unexpected call to ModifyPlan")
	}
}

func withPanicUpdate(opts *ResourceOptions[*testModel, testModel]) {
	opts.Update = func(ctx context.Context, client avngen.Client, plan, state, config *testModel) diag.Diagnostics {
		panic("unexpected call to Update")
	}
}

func withPanicValidateConfig(opts *ResourceOptions[*testModel, testModel]) {
	opts.ValidateConfig = func(ctx context.Context, client avngen.Client, config *testModel) diag.Diagnostics {
		panic("unexpected call to ValidateConfig")
	}
}

func withDefaults(opts *ResourceOptions[*testModel, testModel]) {
	s := testResourceSchema()
	opts.TypeName = "aiven_test_resource"
	opts.Schema = func(context.Context) schema.Schema { return s }
	opts.IDFields = []string{"project", "service_name", "database_name"}
}

func newTestAdapter(fs ...func(opts *ResourceOptions[*testModel, testModel])) *resourceAdapter[*testModel, testModel] {
	var opts ResourceOptions[*testModel, testModel]
	for _, f := range fs {
		f(&opts)
	}
	return NewResource(opts).(*resourceAdapter[*testModel, testModel])
}

type testProviderData struct {
	genClient avngen.Client
}

func (p testProviderData) GetClient() *aiven.Client    { return nil }
func (p testProviderData) GetGenClient() avngen.Client { return p.genClient }

type testConfigValidator struct{}

func (testConfigValidator) Description(context.Context) string         { return "" }
func (testConfigValidator) MarkdownDescription(context.Context) string { return "" }
func (testConfigValidator) ValidateResource(context.Context, resource.ValidateConfigRequest, *resource.ValidateConfigResponse) {
}

type diagExp struct {
	Severity diag.Severity
	Summary  string
	Detail   string
}

func diagExpStringConversionError() diagExp {
	return diagExp{
		Severity: diag.SeverityError,
		Summary:  "Value Conversion Error",
		Detail: `An unexpected error was encountered trying to convert tftypes.Value into adapter.testModel. This is always an error in the provider. Please report the following to the provider developer:

cannot reflect tftypes.String into a struct, must be an object`,
	}
}

func diagExpNullValueBuildError(path, targetType, suggestedTypesType, suggestedPointerType string) diagExp {
	return diagExp{
		Severity: diag.SeverityError,
		Summary:  "Value Conversion Error",
		Detail: fmt.Sprintf(`An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:

Received null value, however the target type cannot handle null values. Use the corresponding `+"`types`"+` package type, a pointer type or a custom type that handles null values.

Path: %s
Target Type: %s
Suggested `+"`types`"+` Type: %s
Suggested Pointer Type: %s`,
			path, targetType, suggestedTypesType, suggestedPointerType,
		),
	}
}

func requireDiagnostics(t *testing.T, got diag.Diagnostics, want ...diagExp) {
	t.Helper()

	require.Len(t, got, len(want))

	for i, w := range want {
		d := got[i]

		require.Equal(t, w.Severity, d.Severity(), "diag[%d] severity", i)
		require.Equal(t, w.Summary, d.Summary(), "diag[%d] summary", i)

		require.Equal(t, w.Detail, d.Detail(), "diag[%d] detail", i)
	}
}

// /////////////////////////////////////////////////////////////////////////////

func TestNewLazyResource(t *testing.T) {
	r := NewLazyResource(ResourceOptions[*testModel, testModel]{
		TypeName: "aiven_test_resource",
	})().(*resourceAdapter[*testModel, testModel])

	var response resource.MetadataResponse
	r.Metadata(t.Context(), resource.MetadataRequest{}, &response)
	require.Equal(t, "aiven_test_resource", response.TypeName)
}

func TestResourceAdapter_Configure(t *testing.T) {
	cl, err := avngen.NewClient(avngen.TokenOpt("token1"), avngen.UserAgentOpt("ua"))
	require.NoError(t, err)

	t.Run("It does nothing when there is no provider data", func(t *testing.T) {
		r := &resourceAdapter[*testModel, testModel]{client: cl}
		rsp := &resource.ConfigureResponse{}
		r.Configure(t.Context(), resource.ConfigureRequest{ProviderData: nil}, rsp)
		require.False(t, rsp.Diagnostics.HasError())
		require.Equal(t, cl, r.client)
	})

	t.Run("It returns an error when provider data is invalid", func(t *testing.T) {
		r := &resourceAdapter[*testModel, testModel]{client: cl}
		rsp := &resource.ConfigureResponse{}
		r.Configure(t.Context(), resource.ConfigureRequest{ProviderData: 123}, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExp{
			Severity: diag.SeverityError,
			Summary:  errmsg.SummaryUnexpectedProviderDataType,
			Detail:   "Expected *aiven.Client, got: int. Please report this issue to the provider developers.",
		})
		require.Equal(t, cl, r.client)
	})

	t.Run("It sets client from provider data", func(t *testing.T) {
		cl2, err := avngen.NewClient(avngen.TokenOpt("token2"), avngen.UserAgentOpt("ua"))
		require.NoError(t, err)

		r := &resourceAdapter[*testModel, testModel]{}
		rsp := &resource.ConfigureResponse{}
		r.Configure(t.Context(), resource.ConfigureRequest{ProviderData: testProviderData{genClient: cl2}}, rsp)
		require.False(t, rsp.Diagnostics.HasError())
		require.Equal(t, cl2, r.client)
	})
}

func TestResourceAdapter_Metadata(t *testing.T) {
	a := newTestAdapter(withDefaults)

	var meta resource.MetadataResponse
	a.Metadata(t.Context(), resource.MetadataRequest{}, &meta)
	require.Equal(t, "aiven_test_resource", meta.TypeName)
}

func TestResourceAdapter_Schema(t *testing.T) {
	s := testResourceSchema()
	a := newTestAdapter(withDefaults)

	var schemaRsp resource.SchemaResponse
	a.Schema(t.Context(), resource.SchemaRequest{}, &schemaRsp)
	require.Equal(t, s, schemaRsp.Schema)
}

func TestResourceAdapter_Create(t *testing.T) {
	ctx := t.Context()
	s := testResourceSchema()

	t.Run("It aborts without Create when plan is invalid", func(t *testing.T) {
		a := newTestAdapter(withDefaults)

		badPlan := tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(tftypes.String, "bad")}
		req := resource.CreateRequest{Plan: badPlan, Config: testConfigFromPlan(s, badPlan)}
		rsp := &resource.CreateResponse{State: testNullState(ctx, s)}

		a.Create(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExpStringConversionError())
		require.True(t, rsp.State.Raw.IsNull())
	})

	t.Run("It aborts without Create when create timeout is invalid", func(t *testing.T) {
		a := newTestAdapter(withDefaults)

		model := &testModel{
			Project:      types.StringValue("p"),
			ServiceName:  types.StringValue("s"),
			DatabaseName: types.StringValue("db"),
			Timeouts:     testTimeoutsObject(t, "create", "invalid"),
		}
		plan := testPlanFromModel(t, ctx, s, model)

		req := resource.CreateRequest{Plan: plan, Config: testConfigFromPlan(s, plan)}
		rsp := &resource.CreateResponse{State: testNullState(ctx, s)}

		a.Create(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExp{
			Severity: diag.SeverityError,
			Summary:  "Invalid Timeout Value",
			Detail:   `Failed to parse timeout value for "create": time: invalid duration "invalid"`,
		})
	})

	t.Run("It errors and doesn't set state when Create returns errors", func(t *testing.T) {
		m := &mock.Mock{}
		m.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(diag.Diagnostics{errmsg.FromError("boom", avngen.Error{Status: 500})}).Once()
		t.Cleanup(func() { m.AssertExpectations(t) })

		a := newTestAdapter(withDefaults, mockResourceOptions(m), withRefreshState)

		plan := testPlanFromModel(t, ctx, s, &testModel{Timeouts: testTimeoutsObject(t)})
		req := resource.CreateRequest{Plan: plan, Config: testConfigFromPlan(s, plan)}
		rsp := &resource.CreateResponse{State: testNullState(ctx, s)}

		a.Create(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExp{
			Severity: diag.SeverityError,
			Summary:  "boom",
			Detail:   (avngen.Error{Status: 500}).Error(),
		})
		require.True(t, rsp.State.Raw.IsNull())
	})

	t.Run("It errors and doesn't set state when refresh Read returns errors", func(t *testing.T) {
		m := &mock.Mock{}
		m.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				args.Get(2).(*testModel).ID = types.StringValue("created")
			}).Return((diag.Diagnostics)(nil)).Once().
			On("Read", mock.Anything, mock.Anything, mock.MatchedBy(func(m *testModel) bool {
				return m != nil && m.ID.ValueString() == "created"
			})).
			Return(diag.Diagnostics{errmsg.FromError("boom", avngen.Error{Status: 500})}).Once()
		t.Cleanup(func() { m.AssertExpectations(t) })

		a := newTestAdapter(withDefaults, mockResourceOptions(m), withRefreshState)

		plan := testPlanFromModel(t, ctx, s, &testModel{Timeouts: testTimeoutsObject(t)})
		req := resource.CreateRequest{Plan: plan, Config: testConfigFromPlan(s, plan)}
		rsp := &resource.CreateResponse{State: testNullState(ctx, s)}

		a.Create(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExp{
			Severity: diag.SeverityError,
			Summary:  "boom",
			Detail:   (avngen.Error{Status: 500}).Error(),
		})
		require.True(t, rsp.State.Raw.IsNull())
	})

	t.Run("It calls Create without calling Read when RefreshState is disabled", func(t *testing.T) {
		m := &mock.Mock{}
		m.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				args.Get(2).(*testModel).ID = types.StringValue("created")
			}).Return((diag.Diagnostics)(nil)).Once()
		t.Cleanup(func() { m.AssertExpectations(t) })

		a := newTestAdapter(withDefaults, mockResourceOptions(m))

		model := &testModel{
			Project:      types.StringValue("p"),
			ServiceName:  types.StringValue("s"),
			DatabaseName: types.StringValue("db"),
			Timeouts:     testTimeoutsObject(t),
		}
		plan := testPlanFromModel(t, ctx, s, model)

		req := resource.CreateRequest{
			Plan:   plan,
			Config: testConfigFromPlan(s, plan),
		}
		rsp := &resource.CreateResponse{
			State: testNullState(ctx, s),
		}

		a.Create(ctx, req, rsp)
		require.False(t, rsp.Diagnostics.HasError())

		var got testModel
		diags := rsp.State.Get(ctx, &got)
		require.False(t, diags.HasError())
		require.Equal(t, "created", got.ID.ValueString())
		require.Equal(t, "db", got.DatabaseName.ValueString())
	})

	t.Run("It calls Create and then Read when RefreshState is enabled", func(t *testing.T) {
		m := &mock.Mock{}
		m.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				args.Get(2).(*testModel).ID = types.StringValue("created")
			}).Return((diag.Diagnostics)(nil)).Once().
			On("Read", mock.Anything, mock.Anything, mock.MatchedBy(func(m *testModel) bool {
				return m != nil && m.ID.ValueString() == "created"
			})).
			Run(func(args mock.Arguments) {
				args.Get(2).(*testModel).DatabaseName = types.StringValue("refreshed")
			}).Return((diag.Diagnostics)(nil)).Once()
		t.Cleanup(func() { m.AssertExpectations(t) })

		a := newTestAdapter(withDefaults, mockResourceOptions(m), withRefreshState)

		model := &testModel{
			Project:      types.StringValue("p"),
			ServiceName:  types.StringValue("s"),
			DatabaseName: types.StringValue("db"),
			Timeouts:     testTimeoutsObject(t),
		}
		plan := testPlanFromModel(t, ctx, s, model)

		req := resource.CreateRequest{
			Plan:   plan,
			Config: testConfigFromPlan(s, plan),
		}
		rsp := &resource.CreateResponse{
			State: testNullState(ctx, s),
		}

		a.Create(ctx, req, rsp)
		require.False(t, rsp.Diagnostics.HasError())

		var got testModel
		diags := rsp.State.Get(ctx, &got)
		require.False(t, diags.HasError())
		require.Equal(t, "created", got.ID.ValueString())
		require.Equal(t, "refreshed", got.DatabaseName.ValueString())
	})
}

func TestResourceAdapter_Read(t *testing.T) {
	ctx := t.Context()
	s := testResourceSchema()

	t.Run("It aborts without Read when state is invalid", func(t *testing.T) {
		a := newTestAdapter(withDefaults)

		badState := tfsdk.State{Schema: s, Raw: tftypes.NewValue(tftypes.String, "bad")}
		req := resource.ReadRequest{State: badState}
		rsp := &resource.ReadResponse{State: testNullState(ctx, s)}

		a.Read(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExpStringConversionError())
		require.True(t, rsp.State.Raw.IsNull())
	})

	t.Run("It aborts without Read when read timeout is invalid", func(t *testing.T) {
		a := newTestAdapter(withDefaults)

		stateModel := &testModel{
			ID:           types.StringValue("id"),
			Timeouts:     testTimeoutsObject(t, "read", "invalid"),
			Project:      types.StringValue("p"),
			ServiceName:  types.StringValue("s"),
			DatabaseName: types.StringValue("db"),
		}
		state := testStateFromModel(t, ctx, s, stateModel)

		req := resource.ReadRequest{State: state}
		rsp := &resource.ReadResponse{State: testNullState(ctx, s)}

		a.Read(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExp{
			Severity: diag.SeverityError,
			Summary:  "Invalid Timeout Value",
			Detail:   `Failed to parse timeout value for "read": time: invalid duration "invalid"`,
		})
		require.True(t, rsp.State.Raw.IsNull())
	})

	t.Run("It removes resource from state when Read returns not found and RemoveMissing is enabled", func(t *testing.T) {
		m := &mock.Mock{}
		m.On("Read", mock.Anything, mock.Anything, mock.MatchedBy(func(m *testModel) bool {
			return m != nil && m.ID.ValueString() == "id"
		})).Return(diag.Diagnostics{errmsg.FromError("not found", avngen.Error{Status: 404})}).Once()
		t.Cleanup(func() { m.AssertExpectations(t) })

		a := newTestAdapter(withDefaults, mockResourceOptions(m), withRemoveMissing)

		stateModel := &testModel{
			ID:           types.StringValue("id"),
			Project:      types.StringValue("p"),
			Timeouts:     testTimeoutsObject(t),
			ServiceName:  types.StringValue("s"),
			DatabaseName: types.StringValue("db"),
		}
		state := testStateFromModel(t, ctx, s, stateModel)

		req := resource.ReadRequest{State: state}
		rsp := &resource.ReadResponse{State: testNullState(ctx, s)}

		a.Read(ctx, req, rsp)
		require.False(t, rsp.Diagnostics.HasError())
		require.True(t, rsp.State.Raw.IsNull())
	})

	t.Run("It errors and doesn't set state when Read returns errors", func(t *testing.T) {
		m := &mock.Mock{}
		m.On("Read", mock.Anything, mock.Anything, mock.MatchedBy(func(m *testModel) bool {
			return m != nil && m.ID.ValueString() == "id"
		})).Return(diag.Diagnostics{errmsg.FromError("boom", avngen.Error{Status: 500})}).Once()
		t.Cleanup(func() { m.AssertExpectations(t) })

		a := newTestAdapter(withDefaults, mockResourceOptions(m))

		stateModel := &testModel{
			ID:           types.StringValue("id"),
			Project:      types.StringValue("p"),
			ServiceName:  types.StringValue("s"),
			DatabaseName: types.StringValue("db"),
			Timeouts:     testTimeoutsObject(t),
		}
		state := testStateFromModel(t, ctx, s, stateModel)

		req := resource.ReadRequest{State: state}
		rsp := &resource.ReadResponse{State: testNullState(ctx, s)}

		a.Read(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExp{
			Severity: diag.SeverityError,
			Summary:  "boom",
			Detail:   (avngen.Error{Status: 500}).Error(),
		})
		require.True(t, rsp.State.Raw.IsNull())
	})

	t.Run("It writes state after successful Read", func(t *testing.T) {
		m := &mock.Mock{}
		m.On("Read", mock.Anything, mock.Anything, mock.MatchedBy(func(m *testModel) bool {
			return m != nil && m.ID.ValueString() == "id"
		})).Run(func(args mock.Arguments) {
			args.Get(2).(*testModel).Project = types.StringValue("updated")
		}).Return((diag.Diagnostics)(nil)).Once()
		t.Cleanup(func() { m.AssertExpectations(t) })

		a := newTestAdapter(withDefaults, mockResourceOptions(m))

		stateModel := &testModel{
			ID:           types.StringValue("id"),
			Project:      types.StringValue("p"),
			ServiceName:  types.StringValue("s"),
			DatabaseName: types.StringValue("db"),
			Timeouts:     testTimeoutsObject(t),
		}
		state := testStateFromModel(t, ctx, s, stateModel)

		req := resource.ReadRequest{State: state}
		rsp := &resource.ReadResponse{State: testNullState(ctx, s)}

		a.Read(ctx, req, rsp)
		require.False(t, rsp.Diagnostics.HasError())

		var got testModel
		diags := rsp.State.Get(ctx, &got)
		require.False(t, diags.HasError())
		require.Equal(t, "updated", got.Project.ValueString())
	})
}

func TestResourceAdapter_Update(t *testing.T) {
	ctx := t.Context()
	s := testResourceSchema()

	t.Run("It aborts without Update when plan/state/config is invalid", func(t *testing.T) {
		a := newTestAdapter(withDefaults)

		badPlan := tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(tftypes.String, "bad")}
		req := resource.UpdateRequest{
			Plan:   badPlan,
			State:  testStateFromModel(t, ctx, s, &testModel{Timeouts: testTimeoutsObject(t)}),
			Config: tfsdk.Config{Schema: s, Raw: tftypes.NewValue(tftypes.String, "bad")},
		}
		rsp := &resource.UpdateResponse{State: testNullState(ctx, s)}

		a.Update(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExpStringConversionError())
		require.True(t, rsp.State.Raw.IsNull())
	})

	t.Run("It aborts without Update when update timeout is invalid", func(t *testing.T) {
		a := newTestAdapter(withDefaults, withPanicUpdate)

		plan := testPlanFromModel(t, ctx, s, &testModel{Timeouts: testTimeoutsObject(t, "update", "invalid")})
		req := resource.UpdateRequest{
			Plan:   plan,
			State:  testStateFromModel(t, ctx, s, &testModel{Timeouts: testTimeoutsObject(t)}),
			Config: testConfigFromPlan(s, plan),
		}
		rsp := &resource.UpdateResponse{State: testNullState(ctx, s)}

		a.Update(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExp{
			Severity: diag.SeverityError,
			Summary:  "Invalid Timeout Value",
			Detail:   `Failed to parse timeout value for "update": time: invalid duration "invalid"`,
		})
		require.True(t, rsp.State.Raw.IsNull())
	})

	t.Run("It errors and doesn't set state when Update returns errors", func(t *testing.T) {
		m := &mock.Mock{}
		m.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(diag.Diagnostics{errmsg.FromError("boom", avngen.Error{Status: 500})}).Once()
		t.Cleanup(func() { m.AssertExpectations(t) })

		a := newTestAdapter(withDefaults, mockResourceOptions(m), withRefreshState)

		plan := testPlanFromModel(t, ctx, s, &testModel{Timeouts: testTimeoutsObject(t)})
		req := resource.UpdateRequest{
			Plan:   plan,
			State:  testStateFromModel(t, ctx, s, &testModel{Timeouts: testTimeoutsObject(t)}),
			Config: testConfigFromPlan(s, plan),
		}
		rsp := &resource.UpdateResponse{State: testNullState(ctx, s)}

		a.Update(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExp{
			Severity: diag.SeverityError,
			Summary:  "boom",
			Detail:   (avngen.Error{Status: 500}).Error(),
		})
		require.True(t, rsp.State.Raw.IsNull())
	})

	t.Run("It errors and doesn't set state when refresh Read returns errors", func(t *testing.T) {
		m := &mock.Mock{}
		m.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				args.Get(2).(*testModel).ID = types.StringValue("updated")
			}).Return((diag.Diagnostics)(nil)).Once().
			On("Read", mock.Anything, mock.Anything, mock.MatchedBy(func(m *testModel) bool {
				return m != nil && m.ID.ValueString() == "updated"
			})).Return(diag.Diagnostics{errmsg.FromError("boom", avngen.Error{Status: 500})}).Once()
		t.Cleanup(func() { m.AssertExpectations(t) })

		a := newTestAdapter(withDefaults, mockResourceOptions(m), withRefreshState)

		plan := testPlanFromModel(t, ctx, s, &testModel{Timeouts: testTimeoutsObject(t)})
		req := resource.UpdateRequest{
			Plan:   plan,
			State:  testStateFromModel(t, ctx, s, &testModel{Timeouts: testTimeoutsObject(t)}),
			Config: testConfigFromPlan(s, plan),
		}
		rsp := &resource.UpdateResponse{State: testNullState(ctx, s)}

		a.Update(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExp{
			Severity: diag.SeverityError,
			Summary:  "boom",
			Detail:   (avngen.Error{Status: 500}).Error(),
		})
		require.True(t, rsp.State.Raw.IsNull())
	})

	t.Run("It skips Update hook when Update is nil", func(t *testing.T) {
		a := newTestAdapter()

		planModel := &testModel{Project: types.StringValue("new"), Timeouts: testTimeoutsObject(t)}
		plan := testPlanFromModel(t, ctx, s, planModel)

		stateModel := &testModel{Project: types.StringValue("old"), Timeouts: testTimeoutsObject(t)}
		state := testStateFromModel(t, ctx, s, stateModel)

		req := resource.UpdateRequest{
			Plan:   plan,
			State:  state,
			Config: testConfigFromPlan(s, plan),
		}
		rsp := &resource.UpdateResponse{State: testNullState(ctx, s)}

		a.Update(ctx, req, rsp)
		require.False(t, rsp.Diagnostics.HasError())

		var got testModel
		diags := rsp.State.Get(ctx, &got)
		require.False(t, diags.HasError())
		require.Equal(t, "new", got.Project.ValueString())
	})

	t.Run("It calls Update and then Read when RefreshState is enabled", func(t *testing.T) {
		m := &mock.Mock{}
		m.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				args.Get(2).(*testModel).ID = types.StringValue("updated")
			}).Return((diag.Diagnostics)(nil)).Once().
			On("Read", mock.Anything, mock.Anything, mock.MatchedBy(func(m *testModel) bool {
				return m != nil && m.ID.ValueString() == "updated"
			})).Run(func(args mock.Arguments) {
			args.Get(2).(*testModel).ServiceName = types.StringValue("refreshed")
		}).Return((diag.Diagnostics)(nil)).Once()
		t.Cleanup(func() { m.AssertExpectations(t) })

		a := newTestAdapter(withDefaults, mockResourceOptions(m), withRefreshState)

		planModel := &testModel{Project: types.StringValue("p"), Timeouts: testTimeoutsObject(t)}
		plan := testPlanFromModel(t, ctx, s, planModel)

		stateModel := &testModel{Project: types.StringValue("p"), Timeouts: testTimeoutsObject(t)}
		state := testStateFromModel(t, ctx, s, stateModel)

		req := resource.UpdateRequest{
			Plan:   plan,
			State:  state,
			Config: testConfigFromPlan(s, plan),
		}
		rsp := &resource.UpdateResponse{State: testNullState(ctx, s)}

		a.Update(ctx, req, rsp)
		require.False(t, rsp.Diagnostics.HasError())

		var got testModel
		diags := rsp.State.Get(ctx, &got)
		require.False(t, diags.HasError())
		require.Equal(t, "updated", got.ID.ValueString())
		require.Equal(t, "refreshed", got.ServiceName.ValueString())
	})
}

func TestResourceAdapter_Delete(t *testing.T) {
	ctx := t.Context()
	s := testResourceSchema()

	t.Run("It aborts without Delete when state is invalid", func(t *testing.T) {
		a := newTestAdapter(withDefaults)

		badState := tfsdk.State{Schema: s, Raw: tftypes.NewValue(tftypes.String, "bad")}
		req := resource.DeleteRequest{State: badState}
		rsp := &resource.DeleteResponse{State: testNullState(ctx, s)}

		a.Delete(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExpStringConversionError())
	})

	t.Run("It aborts without Delete when delete timeout is invalid", func(t *testing.T) {
		a := newTestAdapter(withDefaults)

		stateModel := &testModel{ID: types.StringValue("id"), Timeouts: testTimeoutsObject(t, "delete", "invalid")}
		req := resource.DeleteRequest{State: testStateFromModel(t, ctx, s, stateModel)}
		rsp := &resource.DeleteResponse{State: testNullState(ctx, s)}

		a.Delete(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExp{
			Severity: diag.SeverityError,
			Summary:  "Invalid Timeout Value",
			Detail:   `Failed to parse timeout value for "delete": time: invalid duration "invalid"`,
		})
	})

	t.Run("It ignores not found errors from Delete", func(t *testing.T) {
		m := &mock.Mock{}
		m.On("Delete", mock.Anything, mock.Anything, mock.MatchedBy(func(m *testModel) bool {
			return m != nil && m.ID.ValueString() == "id"
		})).Return(diag.Diagnostics{errmsg.FromError("not found", avngen.Error{Status: 404})}).Once()
		t.Cleanup(func() { m.AssertExpectations(t) })

		a := newTestAdapter(withDefaults, mockResourceOptions(m))

		stateModel := &testModel{ID: types.StringValue("id"), Timeouts: testTimeoutsObject(t)}
		req := resource.DeleteRequest{State: testStateFromModel(t, ctx, s, stateModel)}
		rsp := &resource.DeleteResponse{State: testNullState(ctx, s)}

		a.Delete(ctx, req, rsp)
		require.False(t, rsp.Diagnostics.HasError())
	})

	t.Run("It keeps other errors from Delete", func(t *testing.T) {
		m := &mock.Mock{}
		m.On("Delete", mock.Anything, mock.Anything, mock.MatchedBy(func(m *testModel) bool {
			return m != nil && m.ID.ValueString() == "id"
		})).Return(diag.Diagnostics{errmsg.FromError("boom", avngen.Error{Status: 500})}).Once()
		t.Cleanup(func() { m.AssertExpectations(t) })

		a := newTestAdapter(withDefaults, mockResourceOptions(m))

		stateModel := &testModel{ID: types.StringValue("id"), Timeouts: testTimeoutsObject(t)}
		req := resource.DeleteRequest{State: testStateFromModel(t, ctx, s, stateModel)}
		rsp := &resource.DeleteResponse{State: testNullState(ctx, s)}

		a.Delete(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExp{
			Severity: diag.SeverityError,
			Summary:  "boom",
			Detail:   (avngen.Error{Status: 500}).Error(),
		})
	})
}

func TestResourceAdapter_ImportState(t *testing.T) {
	ctx := t.Context()
	s := testResourceSchema()

	t.Run("It sets ID fields from import identifier", func(t *testing.T) {
		a := newTestAdapter(withDefaults)

		req := resource.ImportStateRequest{ID: "p/s/db"}
		rsp := &resource.ImportStateResponse{
			State: testNullState(ctx, s),
		}

		a.ImportState(ctx, req, rsp)
		require.False(t, rsp.Diagnostics.HasError())

		var got testModel
		diags := rsp.State.Get(ctx, &got)
		require.False(t, diags.HasError())
		require.Equal(t, "p", got.Project.ValueString())
		require.Equal(t, "s", got.ServiceName.ValueString())
		require.Equal(t, "db", got.DatabaseName.ValueString())
	})

	t.Run("It errors and shows expected identifier format when import identifier is invalid", func(t *testing.T) {
		testcases := []struct {
			name string
			id   string
		}{
			{name: "Identifier has two segments", id: "p/s"},
			{name: "Identifier has one segment", id: "broken"},
		}

		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				req := resource.ImportStateRequest{ID: tc.id}
				rsp := &resource.ImportStateResponse{
					State: testNullState(ctx, s),
				}

				a := newTestAdapter(withDefaults)
				a.ImportState(ctx, req, rsp)
				requireDiagnostics(t, rsp.Diagnostics, diagExp{
					Severity: diag.SeverityError,
					Summary:  "Unexpected Read Identifier",
					Detail:   `Expected import identifier with format: "project/service_name/database_name". Got: "` + tc.id + `"`,
				})
			})
		}
	})
}

func TestResourceAdapter_ConfigValidators(t *testing.T) {
	t.Run("It returns nil when ConfigValidators hook is nil", func(t *testing.T) {
		a := newTestAdapter()
		require.Nil(t, a.ConfigValidators(t.Context()))
	})

	t.Run("It returns validators from ConfigValidators hook", func(t *testing.T) {
		m := &mock.Mock{}
		m.On("ConfigValidators", mock.Anything, mock.Anything).
			Return([]resource.ConfigValidator{testConfigValidator{}}).Once()
		t.Cleanup(func() { m.AssertExpectations(t) })

		a := newTestAdapter(withDefaults, mockResourceOptions(m))

		validators := a.ConfigValidators(t.Context())
		require.Len(t, validators, 1)
	})
}

func TestResourceAdapter_ValidateConfig(t *testing.T) {
	ctx := t.Context()
	s := testResourceSchema()

	t.Run("It errors when beta resource is not enabled", func(t *testing.T) {
		t.Setenv("PROVIDER_AIVEN_ENABLE_BETA", "")

		a := newTestAdapter(withDefaults, withBeta)

		cfgModel := &testModel{Timeouts: testTimeoutsObject(t)}
		plan := testPlanFromModel(t, ctx, s, cfgModel)
		req := resource.ValidateConfigRequest{Config: testConfigFromPlan(s, plan)}
		rsp := &resource.ValidateConfigResponse{}

		a.ValidateConfig(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExp{
			Severity: diag.SeverityError,
			Summary:  "Beta Resource Not Enabled",
			Detail:   "The `aiven_test_resource` resource is in beta. Set the `PROVIDER_AIVEN_ENABLE_BETA` environment variable to enable.",
		})
	})

	t.Run("It does nothing when ValidateConfig hook is nil", func(t *testing.T) {
		a := newTestAdapter()

		cfgModel := &testModel{Timeouts: testTimeoutsObject(t)}
		plan := testPlanFromModel(t, ctx, s, cfgModel)
		req := resource.ValidateConfigRequest{Config: testConfigFromPlan(s, plan)}
		rsp := &resource.ValidateConfigResponse{}

		a.ValidateConfig(ctx, req, rsp)
		require.False(t, rsp.Diagnostics.HasError())
	})

	t.Run("It aborts without ValidateConfig when config is invalid", func(t *testing.T) {
		a := newTestAdapter(withDefaults, withPanicValidateConfig)

		req := resource.ValidateConfigRequest{
			Config: tfsdk.Config{Schema: s, Raw: tftypes.NewValue(tftypes.String, "bad")},
		}
		rsp := &resource.ValidateConfigResponse{}

		a.ValidateConfig(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExpStringConversionError())
	})

	t.Run("It aborts without ValidateConfig when read timeout is invalid", func(t *testing.T) {
		a := newTestAdapter(withDefaults, withPanicValidateConfig)

		cfgModel := &testModel{Timeouts: testTimeoutsObject(t, "read", "invalid")}
		plan := testPlanFromModel(t, ctx, s, cfgModel)
		req := resource.ValidateConfigRequest{Config: testConfigFromPlan(s, plan)}
		rsp := &resource.ValidateConfigResponse{}

		a.ValidateConfig(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExp{
			Severity: diag.SeverityError,
			Summary:  "Invalid Timeout Value",
			Detail:   `Failed to parse timeout value for "read": time: invalid duration "invalid"`,
		})
	})

	t.Run("It returns errors from ValidateConfig hook", func(t *testing.T) {
		m := &mock.Mock{}
		m.On("ValidateConfig", mock.Anything, mock.Anything, mock.Anything).
			Return(diag.Diagnostics{errmsg.FromError("boom", avngen.Error{Status: 500})}).Once()
		t.Cleanup(func() { m.AssertExpectations(t) })

		a := newTestAdapter(withDefaults, mockResourceOptions(m))

		cfgModel := &testModel{Timeouts: testTimeoutsObject(t)}
		plan := testPlanFromModel(t, ctx, s, cfgModel)
		req := resource.ValidateConfigRequest{Config: testConfigFromPlan(s, plan)}
		rsp := &resource.ValidateConfigResponse{}

		a.ValidateConfig(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExp{
			Severity: diag.SeverityError,
			Summary:  "boom",
			Detail:   (avngen.Error{Status: 500}).Error(),
		})
	})

	t.Run("It calls ValidateConfig hook when config is valid", func(t *testing.T) {
		m := &mock.Mock{}
		m.On("ValidateConfig", mock.Anything, mock.Anything, mock.Anything).
			Return((diag.Diagnostics)(nil)).Once()
		t.Cleanup(func() { m.AssertExpectations(t) })

		a := newTestAdapter(withDefaults, mockResourceOptions(m))

		cfgModel := &testModel{Timeouts: testTimeoutsObject(t)}
		plan := testPlanFromModel(t, ctx, s, cfgModel)
		req := resource.ValidateConfigRequest{Config: testConfigFromPlan(s, plan)}
		rsp := &resource.ValidateConfigResponse{}

		a.ValidateConfig(ctx, req, rsp)
		require.False(t, rsp.Diagnostics.HasError())
	})
}

func TestResourceAdapter_ModifyPlan(t *testing.T) {
	ctx := t.Context()
	s := testResourceSchema()

	t.Run("It errors when termination_protection cannot be read from state", func(t *testing.T) {
		a := newTestAdapter(withDefaults, withTerminationProtection)

		destroyPlan := tfsdk.Plan{Schema: s, Raw: testNullTerraformValue(ctx, s)}
		badState := tfsdk.State{Schema: s, Raw: tftypes.NewValue(tftypes.String, "bad")}
		req := resource.ModifyPlanRequest{
			Plan:  destroyPlan,
			State: badState,
		}
		rsp := &resource.ModifyPlanResponse{Plan: destroyPlan}

		a.ModifyPlan(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExpNullValueBuildError("termination_protection", "bool", "basetypes.BoolValue", "*bool"))
	})

	t.Run("It blocks deletion when termination protection is enabled", func(t *testing.T) {
		a := newTestAdapter(withDefaults, withTerminationProtection)

		destroyPlan := tfsdk.Plan{Schema: s, Raw: testNullTerraformValue(ctx, s)}
		destroyState := testStateFromModel(t, ctx, s, &testModel{
			TerminationProtection: types.BoolValue(true),
			Timeouts:              testTimeoutsObject(t),
		})
		req := resource.ModifyPlanRequest{
			Plan:  destroyPlan,
			State: destroyState,
		}
		rsp := &resource.ModifyPlanResponse{Plan: destroyPlan}

		a.ModifyPlan(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExp{
			Severity: diag.SeverityError,
			Summary:  errmsg.SummaryErrorDeletingResource,
			Detail:   "The resource `aiven_test_resource` has termination protection enabled and cannot be deleted.",
		})
	})

	t.Run("It does nothing when ModifyPlan hook is nil", func(t *testing.T) {
		a := newTestAdapter()

		plan := testPlanFromModel(t, ctx, s, &testModel{Timeouts: testTimeoutsObject(t)})
		req := resource.ModifyPlanRequest{
			Plan:   plan,
			State:  testNullState(ctx, s),
			Config: testConfigFromPlan(s, plan),
		}
		rsp := &resource.ModifyPlanResponse{Plan: plan}

		a.ModifyPlan(ctx, req, rsp)
		require.False(t, rsp.Diagnostics.HasError())
	})

	t.Run("It doesn't call ModifyPlan hook for destroy plans", func(t *testing.T) {
		a := newTestAdapter(withDefaults, withPanicModifyPlan)

		destroyPlan := tfsdk.Plan{Schema: s, Raw: testNullTerraformValue(ctx, s)}
		destroyState := testStateFromModel(t, ctx, s, &testModel{
			TerminationProtection: types.BoolValue(false),
			Timeouts:              testTimeoutsObject(t),
		})
		req := resource.ModifyPlanRequest{
			Plan:   destroyPlan,
			State:  destroyState,
			Config: testConfigFromPlan(s, destroyPlan),
		}
		rsp := &resource.ModifyPlanResponse{Plan: destroyPlan}

		a.ModifyPlan(ctx, req, rsp)
		require.False(t, rsp.Diagnostics.HasError())
	})

	t.Run("It aborts without ModifyPlan when plan is invalid", func(t *testing.T) {
		a := newTestAdapter(withDefaults, withPanicModifyPlan)

		badPlan := tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(tftypes.String, "bad")}
		req := resource.ModifyPlanRequest{
			Plan:   badPlan,
			State:  testNullState(ctx, s),
			Config: tfsdk.Config{Schema: s, Raw: testNullTerraformValue(ctx, s)},
		}
		rsp := &resource.ModifyPlanResponse{Plan: badPlan}

		a.ModifyPlan(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics,
			diagExpStringConversionError(),
			diagExpNullValueBuildError("", "adapter.testModel", "basetypes.ObjectValue", "*adapter.testModel"),
		)
	})

	t.Run("It aborts without ModifyPlan when config is invalid", func(t *testing.T) {
		a := newTestAdapter(withDefaults, withPanicModifyPlan)

		plan := testPlanFromModel(t, ctx, s, &testModel{Timeouts: testTimeoutsObject(t)})
		req := resource.ModifyPlanRequest{
			Plan:   plan,
			State:  testNullState(ctx, s),
			Config: tfsdk.Config{Schema: s, Raw: tftypes.NewValue(tftypes.String, "bad")},
		}
		rsp := &resource.ModifyPlanResponse{Plan: plan}

		a.ModifyPlan(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExpStringConversionError())
	})

	t.Run("It aborts without ModifyPlan when state is invalid", func(t *testing.T) {
		a := newTestAdapter(withDefaults, withPanicModifyPlan)

		plan := testPlanFromModel(t, ctx, s, &testModel{Timeouts: testTimeoutsObject(t)})
		req := resource.ModifyPlanRequest{
			Plan:   plan,
			State:  tfsdk.State{Schema: s, Raw: tftypes.NewValue(tftypes.String, "bad")},
			Config: testConfigFromPlan(s, plan),
		}
		rsp := &resource.ModifyPlanResponse{Plan: plan}

		a.ModifyPlan(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExpStringConversionError())
	})

	t.Run("It aborts without ModifyPlan when read timeout is invalid", func(t *testing.T) {
		a := newTestAdapter(withDefaults, withPanicModifyPlan)

		plan := testPlanFromModel(t, ctx, s, &testModel{Timeouts: testTimeoutsObject(t, "read", "invalid")})
		req := resource.ModifyPlanRequest{
			Plan:   plan,
			State:  testNullState(ctx, s),
			Config: testConfigFromPlan(s, plan),
		}
		rsp := &resource.ModifyPlanResponse{Plan: plan}

		a.ModifyPlan(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExp{
			Severity: diag.SeverityError,
			Summary:  "Invalid Timeout Value",
			Detail:   `Failed to parse timeout value for "read": time: invalid duration "invalid"`,
		})
	})

	t.Run("It returns errors from ModifyPlan hook", func(t *testing.T) {
		m := &mock.Mock{}
		m.On("ModifyPlan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(diag.Diagnostics{errmsg.FromError("boom", avngen.Error{Status: 500})}).Once()
		t.Cleanup(func() { m.AssertExpectations(t) })

		a := newTestAdapter(withDefaults, mockResourceOptions(m))

		plan := testPlanFromModel(t, ctx, s, &testModel{Timeouts: testTimeoutsObject(t)})
		req := resource.ModifyPlanRequest{
			Plan:   plan,
			State:  testNullState(ctx, s),
			Config: testConfigFromPlan(s, plan),
		}
		rsp := &resource.ModifyPlanResponse{Plan: plan}

		a.ModifyPlan(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExp{
			Severity: diag.SeverityError,
			Summary:  "boom",
			Detail:   (avngen.Error{Status: 500}).Error(),
		})
	})

	t.Run("It passes nil state to ModifyPlan during create planning", func(t *testing.T) {
		m := &mock.Mock{}
		m.On("ModifyPlan", mock.Anything, mock.Anything, mock.Anything, mock.MatchedBy(func(m *testModel) bool {
			return m == nil
		}), mock.Anything).Run(func(args mock.Arguments) {
			args.Get(2).(*testModel).DatabaseName = types.StringValue("modified")
		}).Return((diag.Diagnostics)(nil)).Once()
		t.Cleanup(func() { m.AssertExpectations(t) })

		a := newTestAdapter(withDefaults, mockResourceOptions(m))

		plan := testPlanFromModel(t, ctx, s, &testModel{DatabaseName: types.StringValue("db"), Timeouts: testTimeoutsObject(t)})
		req := resource.ModifyPlanRequest{
			Plan:   plan,
			State:  testNullState(ctx, s), // create plan has no state yet
			Config: testConfigFromPlan(s, plan),
		}
		rsp := &resource.ModifyPlanResponse{Plan: plan}

		a.ModifyPlan(ctx, req, rsp)
		require.False(t, rsp.Diagnostics.HasError())

		var got testModel
		diags := rsp.Plan.Get(ctx, &got)
		require.False(t, diags.HasError())
		require.Equal(t, "modified", got.DatabaseName.ValueString())
	})

	t.Run("It passes state to ModifyPlan during update planning", func(t *testing.T) {
		m := &mock.Mock{}
		m.On("ModifyPlan", mock.Anything, mock.Anything, mock.Anything, mock.MatchedBy(func(m *testModel) bool {
			return m != nil && m.DatabaseName.ValueString() == "from-state"
		}), mock.Anything).Run(func(args mock.Arguments) {
			args.Get(2).(*testModel).DatabaseName = types.StringValue("modified")
		}).Return((diag.Diagnostics)(nil)).Once()
		t.Cleanup(func() { m.AssertExpectations(t) })

		a := newTestAdapter(withDefaults, mockResourceOptions(m))

		plan := testPlanFromModel(t, ctx, s, &testModel{DatabaseName: types.StringValue("from-plan"), Timeouts: testTimeoutsObject(t)})
		state := testStateFromModel(t, ctx, s, &testModel{DatabaseName: types.StringValue("from-state"), Timeouts: testTimeoutsObject(t)})
		req := resource.ModifyPlanRequest{
			Plan:   plan,
			State:  state,
			Config: testConfigFromPlan(s, plan),
		}
		rsp := &resource.ModifyPlanResponse{Plan: plan}

		a.ModifyPlan(ctx, req, rsp)
		require.False(t, rsp.Diagnostics.HasError())

		var got testModel
		diags := rsp.Plan.Get(ctx, &got)
		require.False(t, diags.HasError())
		require.Equal(t, "modified", got.DatabaseName.ValueString())
	})
}
