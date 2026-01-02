package adapter

import (
	"context"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

func testDataSourceSchema() dsschema.Schema {
	return dsschema.Schema{
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed: true,
			},
			"project": dsschema.StringAttribute{
				Optional: true,
			},
			"service_name": dsschema.StringAttribute{
				Optional: true,
			},
			"database_name": dsschema.StringAttribute{
				Optional: true,
			},
			"termination_protection": dsschema.BoolAttribute{
				Optional: true,
			},
			"timeouts": dsschema.ObjectAttribute{
				Optional:       true,
				AttributeTypes: testTimeoutAttrTypes,
			},
		},
	}
}

func testNullTerraformValueDataSource(ctx context.Context, s dsschema.Schema) tftypes.Value {
	return tftypes.NewValue(s.Type().TerraformType(ctx), nil)
}

func testConfigFromModelDataSource(t *testing.T, ctx context.Context, s dsschema.Schema, m *testModel) tfsdk.Config {
	t.Helper()

	state := testNullStateDataSource(ctx, s)
	diags := state.Set(ctx, m)
	require.False(t, diags.HasError())
	return tfsdk.Config{Schema: s, Raw: state.Raw}
}

func testNullStateDataSource(ctx context.Context, s dsschema.Schema) tfsdk.State {
	return tfsdk.State{
		Schema: s,
		Raw:    testNullTerraformValueDataSource(ctx, s),
	}
}

func mockDataSourceOptions(m *mock.Mock) func(opts *DataSourceOptions[*testModel, testModel]) {
	return func(opts *DataSourceOptions[*testModel, testModel]) {
		opts.Read = func(ctx context.Context, cl avngen.Client, state *testModel) diag.Diagnostics {
			args := m.MethodCalled("Read", ctx, cl, state)
			return args.Get(0).(diag.Diagnostics)
		}
		opts.ConfigValidators = func(ctx context.Context, cl avngen.Client) []datasource.ConfigValidator {
			args := m.MethodCalled("ConfigValidators", ctx, cl)
			return args.Get(0).([]datasource.ConfigValidator)
		}
		opts.ValidateConfig = func(ctx context.Context, cl avngen.Client, config *testModel) diag.Diagnostics {
			args := m.MethodCalled("ValidateConfig", ctx, cl, config)
			return args.Get(0).(diag.Diagnostics)
		}
	}
}

func withDataSourceDefaults(opts *DataSourceOptions[*testModel, testModel]) {
	s := testDataSourceSchema()
	opts.TypeName = "aiven_test_datasource"
	opts.Schema = func(context.Context) dsschema.Schema { return s }
}

func withDataSourceBeta(opts *DataSourceOptions[*testModel, testModel]) {
	opts.Beta = true
}

func withPanicDataSourceRead(opts *DataSourceOptions[*testModel, testModel]) {
	opts.Read = func(context.Context, avngen.Client, *testModel) diag.Diagnostics {
		panic("unexpected call to Read")
	}
}

func withPanicDataSourceValidateConfig(opts *DataSourceOptions[*testModel, testModel]) {
	opts.ValidateConfig = func(context.Context, avngen.Client, *testModel) diag.Diagnostics {
		panic("unexpected call to ValidateConfig")
	}
}

func newTestDataSourceAdapter(fs ...func(opts *DataSourceOptions[*testModel, testModel])) *datasourceAdapter[*testModel, testModel] {
	var opts DataSourceOptions[*testModel, testModel]
	for _, f := range fs {
		f(&opts)
	}

	return NewDataSource(opts).(*datasourceAdapter[*testModel, testModel])
}

type testDataSourceConfigValidator struct{}

func (testDataSourceConfigValidator) Description(context.Context) string         { return "" }
func (testDataSourceConfigValidator) MarkdownDescription(context.Context) string { return "" }
func (testDataSourceConfigValidator) ValidateDataSource(context.Context, datasource.ValidateConfigRequest, *datasource.ValidateConfigResponse) {
}

///////////////////////////////////////////////////////////////////////////////

func TestNewLazyDataSource(t *testing.T) {
	a := NewLazyDataSource(DataSourceOptions[*testModel, testModel]{
		TypeName: "aiven_test_datasource",
	})().(*datasourceAdapter[*testModel, testModel])

	var rsp datasource.MetadataResponse
	a.Metadata(t.Context(), datasource.MetadataRequest{}, &rsp)
	require.Equal(t, "aiven_test_datasource", rsp.TypeName)
}

func TestDataSourceAdapter_Configure(t *testing.T) {
	cl1, err := avngen.NewClient(avngen.TokenOpt("token1"), avngen.UserAgentOpt("ua"))
	require.NoError(t, err)

	t.Run("It does nothing when there is no provider data", func(t *testing.T) {
		a := &datasourceAdapter[*testModel, testModel]{client: cl1}
		rsp := &datasource.ConfigureResponse{}
		a.Configure(t.Context(), datasource.ConfigureRequest{ProviderData: nil}, rsp)
		require.False(t, rsp.Diagnostics.HasError())
		require.Equal(t, cl1, a.client)
	})

	t.Run("It returns an error when provider data is invalid", func(t *testing.T) {
		a := &datasourceAdapter[*testModel, testModel]{client: cl1}
		rsp := &datasource.ConfigureResponse{}
		a.Configure(t.Context(), datasource.ConfigureRequest{ProviderData: 123}, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExp{
			Severity: diag.SeverityError,
			Summary:  errmsg.SummaryUnexpectedProviderDataType,
			Detail:   "Expected *aiven.Client, got: int. Please report this issue to the provider developers.",
		})
		require.Equal(t, cl1, a.client)
	})

	t.Run("It sets client from provider data", func(t *testing.T) {
		cl2, err := avngen.NewClient(avngen.TokenOpt("token2"), avngen.UserAgentOpt("ua"))
		require.NoError(t, err)

		a := &datasourceAdapter[*testModel, testModel]{}
		rsp := &datasource.ConfigureResponse{}
		a.Configure(t.Context(), datasource.ConfigureRequest{ProviderData: testProviderData{genClient: cl2}}, rsp)
		require.False(t, rsp.Diagnostics.HasError())
		require.Equal(t, cl2, a.client)
	})
}

func TestDataSourceAdapter_Metadata(t *testing.T) {
	a := newTestDataSourceAdapter(withDataSourceDefaults)

	var meta datasource.MetadataResponse
	a.Metadata(t.Context(), datasource.MetadataRequest{}, &meta)
	require.Equal(t, "aiven_test_datasource", meta.TypeName)
}

func TestDataSourceAdapter_Schema(t *testing.T) {
	s := testDataSourceSchema()
	a := newTestDataSourceAdapter(withDataSourceDefaults)

	var rsp datasource.SchemaResponse
	a.Schema(t.Context(), datasource.SchemaRequest{}, &rsp)
	require.Equal(t, s, rsp.Schema)
}

func TestDataSourceAdapter_Read(t *testing.T) {
	ctx := t.Context()
	s := testDataSourceSchema()

	t.Run("It aborts without Read when config is invalid", func(t *testing.T) {
		a := newTestDataSourceAdapter(withDataSourceDefaults, withPanicDataSourceRead)

		req := datasource.ReadRequest{
			Config: tfsdk.Config{Schema: s, Raw: tftypes.NewValue(tftypes.String, "bad")},
		}
		rsp := &datasource.ReadResponse{State: testNullStateDataSource(ctx, s)}

		a.Read(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExpStringConversionError())
		require.True(t, rsp.State.Raw.IsNull())
	})

	t.Run("It aborts without Read when read timeout is invalid", func(t *testing.T) {
		a := newTestDataSourceAdapter(withDataSourceDefaults, withPanicDataSourceRead)

		cfg := testConfigFromModelDataSource(t, ctx, s, &testModel{Timeouts: testTimeoutsObject(t, "read", "invalid")})
		req := datasource.ReadRequest{Config: cfg}
		rsp := &datasource.ReadResponse{State: testNullStateDataSource(ctx, s)}

		a.Read(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExp{
			Severity: diag.SeverityError,
			Summary:  "Invalid Timeout Value",
			Detail:   `Failed to parse timeout value for "read": time: invalid duration "invalid"`,
		})
		require.True(t, rsp.State.Raw.IsNull())
	})

	t.Run("It errors and does not set state when Read returns errors", func(t *testing.T) {
		m := &mock.Mock{}
		m.On("Read", mock.Anything, mock.Anything, mock.Anything).
			Return(diag.Diagnostics{errmsg.FromError("boom", avngen.Error{Status: 500})}).Once()
		t.Cleanup(func() { m.AssertExpectations(t) })

		a := newTestDataSourceAdapter(withDataSourceDefaults, mockDataSourceOptions(m))

		cfg := testConfigFromModelDataSource(t, ctx, s, &testModel{Timeouts: testTimeoutsObject(t)})
		req := datasource.ReadRequest{Config: cfg}
		rsp := &datasource.ReadResponse{State: testNullStateDataSource(ctx, s)}

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
		m.On("Read", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			args.Get(2).(*testModel).ID = types.StringValue("id")
			args.Get(2).(*testModel).Project = types.StringValue("updated")
		}).Return((diag.Diagnostics)(nil)).Once()
		t.Cleanup(func() { m.AssertExpectations(t) })

		a := newTestDataSourceAdapter(withDataSourceDefaults, mockDataSourceOptions(m))

		cfg := testConfigFromModelDataSource(t, ctx, s, &testModel{
			Project:  types.StringValue("p"),
			Timeouts: testTimeoutsObject(t),
		})
		req := datasource.ReadRequest{Config: cfg}
		rsp := &datasource.ReadResponse{State: testNullStateDataSource(ctx, s)}

		a.Read(ctx, req, rsp)
		require.False(t, rsp.Diagnostics.HasError())

		var got testModel
		diags := rsp.State.Get(ctx, &got)
		require.False(t, diags.HasError())
		require.Equal(t, "id", got.ID.ValueString())
		require.Equal(t, "updated", got.Project.ValueString())
	})
}

func TestDataSourceAdapter_ValidateConfig(t *testing.T) {
	ctx := t.Context()
	s := testDataSourceSchema()

	t.Run("It errors when beta datasource is not enabled", func(t *testing.T) {
		t.Setenv("PROVIDER_AIVEN_ENABLE_BETA", "")

		a := newTestDataSourceAdapter(withDataSourceDefaults, withDataSourceBeta, withPanicDataSourceValidateConfig)

		cfg := testConfigFromModelDataSource(t, ctx, s, &testModel{Timeouts: testTimeoutsObject(t)})
		req := datasource.ValidateConfigRequest{Config: cfg}
		rsp := &datasource.ValidateConfigResponse{}

		a.ValidateConfig(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExp{
			Severity: diag.SeverityError,
			Summary:  "Beta DataSource Not Enabled",
			Detail:   "The `aiven_test_datasource` datasource is in beta. Set the `PROVIDER_AIVEN_ENABLE_BETA=true` environment variable to enable.",
		})
	})

	t.Run("It does nothing when ValidateConfig hook is nil", func(t *testing.T) {
		a := newTestDataSourceAdapter()

		cfg := testConfigFromModelDataSource(t, ctx, s, &testModel{Timeouts: testTimeoutsObject(t)})
		req := datasource.ValidateConfigRequest{Config: cfg}
		rsp := &datasource.ValidateConfigResponse{}

		a.ValidateConfig(ctx, req, rsp)
		require.False(t, rsp.Diagnostics.HasError())
	})

	t.Run("It aborts without ValidateConfig when config is invalid", func(t *testing.T) {
		a := newTestDataSourceAdapter(withDataSourceDefaults, withPanicDataSourceValidateConfig)

		req := datasource.ValidateConfigRequest{
			Config: tfsdk.Config{Schema: s, Raw: tftypes.NewValue(tftypes.String, "bad")},
		}
		rsp := &datasource.ValidateConfigResponse{}

		a.ValidateConfig(ctx, req, rsp)
		requireDiagnostics(t, rsp.Diagnostics, diagExpStringConversionError())
	})

	t.Run("It aborts without ValidateConfig when read timeout is invalid", func(t *testing.T) {
		a := newTestDataSourceAdapter(withDataSourceDefaults, withPanicDataSourceValidateConfig)

		cfg := testConfigFromModelDataSource(t, ctx, s, &testModel{Timeouts: testTimeoutsObject(t, "read", "invalid")})
		req := datasource.ValidateConfigRequest{Config: cfg}
		rsp := &datasource.ValidateConfigResponse{}

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

		a := newTestDataSourceAdapter(withDataSourceDefaults, mockDataSourceOptions(m))

		cfg := testConfigFromModelDataSource(t, ctx, s, &testModel{Timeouts: testTimeoutsObject(t)})
		req := datasource.ValidateConfigRequest{Config: cfg}
		rsp := &datasource.ValidateConfigResponse{}

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

		a := newTestDataSourceAdapter(withDataSourceDefaults, mockDataSourceOptions(m))

		cfg := testConfigFromModelDataSource(t, ctx, s, &testModel{Timeouts: testTimeoutsObject(t)})
		req := datasource.ValidateConfigRequest{Config: cfg}
		rsp := &datasource.ValidateConfigResponse{}

		a.ValidateConfig(ctx, req, rsp)
		require.False(t, rsp.Diagnostics.HasError())
	})
}

func TestDataSourceAdapter_ConfigValidators(t *testing.T) {
	t.Run("It returns nil when ConfigValidators hook is nil", func(t *testing.T) {
		a := newTestDataSourceAdapter()
		require.Nil(t, a.ConfigValidators(t.Context()))
	})

	t.Run("It returns validators from ConfigValidators hook", func(t *testing.T) {
		m := &mock.Mock{}
		m.On("ConfigValidators", mock.Anything, mock.Anything).
			Return([]datasource.ConfigValidator{testDataSourceConfigValidator{}}).Once()
		t.Cleanup(func() { m.AssertExpectations(t) })

		a := newTestDataSourceAdapter(withDataSourceDefaults, mockDataSourceOptions(m))

		validators := a.ConfigValidators(t.Context())
		require.Len(t, validators, 1)
	})
}
