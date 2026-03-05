package adapter

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/providerdata"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

type DataSourceOptions struct {
	// TypeName is the name of resource,
	// for instance, "aiven_organization_address"
	TypeName       string
	Schema         func(ctx context.Context) schema.Schema
	SchemaInternal *Schema

	// Indicates whether the datasource is in beta.
	// Requires the `PROVIDER_AIVEN_ENABLE_BETA` environment variable to be set.
	Beta bool

	// IDFields are used to build the resource ID.
	// Example: ["project_id", "instance_name"] == "project-123/instance-456"
	IDFields []string

	// Read is the only CRUD operation for datasource.
	Read func(ctx context.Context, client avngen.Client, d ResourceData) error

	// ValidateConfig implements datasource.DataSourceWithValidateConfig.
	// https://developer.hashicorp.com/terraform/plugin/framework/data-sources/validate-configuration#validateconfig-method
	ValidateConfig func(ctx context.Context, client avngen.Client, d ResourceData) error

	// ConfigValidators implements datasource.DataSourceWithConfigValidators.
	// https://developer.hashicorp.com/terraform/plugin/framework/data-sources/validate-configuration#configvalidators-method
	ConfigValidators func(ctx context.Context, client avngen.Client) []datasource.ConfigValidator
}

func NewDataSource(options DataSourceOptions) datasource.DataSource {
	return &datasourceAdapter{
		datasource: options,
	}
}

// NewLazyDataSource creates a lazy resource constructor.
// The provider.Provider.DataSources requires a function that returns a datasource.DataSource.
func NewLazyDataSource(options DataSourceOptions) func() datasource.DataSource {
	return func() datasource.DataSource {
		return NewDataSource(options)
	}
}

var (
	_ datasource.DataSource                     = (*datasourceAdapter)(nil)
	_ datasource.DataSourceWithConfigure        = (*datasourceAdapter)(nil)
	_ datasource.DataSourceWithValidateConfig   = (*datasourceAdapter)(nil)
	_ datasource.DataSourceWithConfigValidators = (*datasourceAdapter)(nil)
)

type datasourceAdapter struct {
	client     avngen.Client
	datasource DataSourceOptions
}

func (a *datasourceAdapter) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	rsp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		// DF calls Configure many times, it might not contain the provider data yet.
		return
	}

	p, diags := providerdata.FromRequest(req.ProviderData)
	if diags.HasError() {
		rsp.Diagnostics.Append(diags...)
		return
	}

	a.client = p.GetGenClient()
}

func (a *datasourceAdapter) Metadata(
	_ context.Context,
	_ datasource.MetadataRequest,
	rsp *datasource.MetadataResponse,
) {
	rsp.TypeName = a.datasource.TypeName
}

func (a *datasourceAdapter) Schema(
	ctx context.Context,
	_ datasource.SchemaRequest,
	rsp *datasource.SchemaResponse,
) {
	rsp.Schema = a.datasource.Schema(ctx)
}

func (a *datasourceAdapter) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	rsp *datasource.ReadResponse,
) {
	diags := &rsp.Diagnostics

	d, err := NewResourceData(a.datasource.SchemaInternal, a.datasource.IDFields, nil, nil, &req.Config)
	if err != nil {
		diags.AddError("failed to create ResourceData", err.Error())
		return
	}

	ctx, cancel, err := withTimeout(ctx, d, timeoutRead)
	if err != nil {
		diags.AddError("failed to read timeouts block", err.Error())
		return
	}
	defer cancel()

	err = a.datasource.Read(ctx, a.client, d)
	if err != nil {
		diags.AddError("failed to read datasource", err.Error())
		return
	}

	rsp.State.Raw = d.tfValue()
}

func (a *datasourceAdapter) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, rsp *datasource.ValidateConfigResponse) {
	if a.datasource.Beta && !util.IsBeta() {
		rsp.Diagnostics.AddError(
			"Beta DataSource Not Enabled",
			fmt.Sprintf("The `%s` datasource is in beta. Set the `%s=true` environment variable to enable.", a.datasource.TypeName, util.AivenEnableBeta),
		)
		return
	}

	if a.datasource.ValidateConfig == nil {
		return
	}

	d, err := NewResourceData(a.datasource.SchemaInternal, a.datasource.IDFields, nil, nil, &req.Config)
	if err != nil {
		rsp.Diagnostics.AddError("failed to create ResourceData", err.Error())
		return
	}

	diags := &rsp.Diagnostics

	// Some datasources might run API calls to validate the configuration.
	ctx, cancel, err := withTimeout(ctx, d, timeoutRead)
	if err != nil {
		diags.AddError("failed to read timeouts block", err.Error())
		return
	}
	defer cancel()

	err = a.datasource.ValidateConfig(ctx, a.client, d)
	if err != nil {
		diags.AddError("failed to validate config", err.Error())
		return
	}
}

func (a *datasourceAdapter) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	if a.datasource.ConfigValidators == nil {
		return nil
	}
	return a.datasource.ConfigValidators(ctx, a.client)
}
