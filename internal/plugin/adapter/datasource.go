package adapter

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/providerdata"
)

type DataSourceOptions[M Model[T], T any] struct {
	// TypeName is the name of resource,
	// for instance, "aiven_organization_address"
	TypeName string
	Schema   func(ctx context.Context) schema.Schema

	// Read is the only CRUD operation for datasource.
	Read func(ctx context.Context, client avngen.Client, state *T) diag.Diagnostics

	// ConfigValidators implements datasource.DataSourceWithConfigValidators.
	// https://developer.hashicorp.com/terraform/plugin/framework/data-sources/validate-configuration#configvalidators-method
	ConfigValidators func(ctx context.Context, client avngen.Client) []datasource.ConfigValidator
}

func NewDataSource[M Model[T], T any](options DataSourceOptions[M, T]) datasource.DataSource {
	return &datasourceAdapter[M, T]{
		datasource: options,
	}
}

var (
	_ datasource.DataSource                     = (*datasourceAdapter[Model[any], any])(nil)
	_ datasource.DataSourceWithConfigure        = (*datasourceAdapter[Model[any], any])(nil)
	_ datasource.DataSourceWithConfigValidators = (*datasourceAdapter[Model[any], any])(nil)
)

type datasourceAdapter[M Model[T], T any] struct {
	client     avngen.Client
	datasource DataSourceOptions[M, T]
}

func (a *datasourceAdapter[M, T]) Configure(
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

func (a *datasourceAdapter[M, T]) Metadata(
	_ context.Context,
	_ datasource.MetadataRequest,
	rsp *datasource.MetadataResponse,
) {
	rsp.TypeName = a.datasource.TypeName
}

func (a *datasourceAdapter[M, T]) Schema(
	ctx context.Context,
	_ datasource.SchemaRequest,
	rsp *datasource.SchemaResponse,
) {
	rsp.Schema = a.datasource.Schema(ctx)
}

func (a *datasourceAdapter[M, T]) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	rsp *datasource.ReadResponse,
) {
	var (
		state = instantiate[M]()
		diags = &rsp.Diagnostics
	)

	diags.Append(req.Config.Get(ctx, state)...)
	if diags.HasError() {
		return
	}

	ctx, cancel, d := withTimeout(ctx, state.TimeoutsObject(), timeoutRead)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	defer cancel()

	diags.Append(a.datasource.Read(ctx, a.client, state.SharedModel())...)
	if diags.HasError() {
		return
	}

	diags.Append(rsp.State.Set(ctx, state)...)
}

func (a *datasourceAdapter[M, T]) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	if a.datasource.ConfigValidators == nil {
		return nil
	}
	return a.datasource.ConfigValidators(ctx, a.client)
}
