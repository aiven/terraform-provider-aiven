package adapter

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/providerdata"
)

// MightyDatasource implements additional datasource methods
type MightyDatasource interface {
	datasource.DataSourceWithConfigure
	datasource.DataSourceWithConfigValidators
}

type datasourceSchema func(context.Context) schema.Schema

func NewDatasource[T any](
	name string,
	view DatView[T],
	newSchema datasourceSchema,
	newDataModel dataModelFactory[T],
) MightyDatasource {
	return &datasourceAdapter[T]{
		name:         name,
		view:         view,
		newSchema:    newSchema,
		newDataModel: newDataModel,
	}
}

type datasourceAdapter[T any] struct {
	// name is the name of the resource or datasource,
	// for instance, "aiven_organization_address"
	name string

	// view implements Read function
	view DatView[T]

	// newSchema returns a new instance of the generated schema.
	newSchema datasourceSchema

	// newDataModel returns a new instance of the generated dataModel.
	newDataModel dataModelFactory[T]
}

func (a *datasourceAdapter[T]) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	rsp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		// TF calls Configure many times, it might not contain the provider data yet.
		return
	}

	p, diags := providerdata.FromRequest(req.ProviderData)
	if diags.HasError() {
		rsp.Diagnostics.Append(diags...)
		return
	}

	a.view.Configure(p.GetGenClient())
}

func (a *datasourceAdapter[T]) Metadata(
	_ context.Context,
	_ datasource.MetadataRequest,
	rsp *datasource.MetadataResponse,
) {
	rsp.TypeName = a.name
}

func (a *datasourceAdapter[T]) Schema(
	ctx context.Context,
	_ datasource.SchemaRequest,
	rsp *datasource.SchemaResponse,
) {
	rsp.Schema = a.newSchema(ctx)
}

func (a *datasourceAdapter[T]) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	rsp *datasource.ReadResponse,
) {
	var (
		state = a.newDataModel()
		diags = &rsp.Diagnostics
	)
	diags.Append(req.Config.Get(ctx, state)...)
	if diags.HasError() {
		return
	}

	diags.Append(a.view.Read(ctx, state.DataModel())...)
	if diags.HasError() {
		return
	}

	diags.Append(rsp.State.Set(ctx, state)...)
}

func (a *datasourceAdapter[T]) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	v, ok := a.view.(DatViewValidators[T])
	if !ok {
		return nil
	}
	return v.DatValidators(ctx)
}
