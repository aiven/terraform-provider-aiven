package controller

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/types"
)

// NewDatasource returns datasource.DataSource but must support Configure
var _ datasource.DataSourceWithConfigure = (*dataController[any])(nil)

type datasourceSchema func(context.Context) schema.Schema

func NewDatasource[T any](
	name string,
	view View[T],
	newSchema datasourceSchema,
	newDataModel dataModelFactory[T],
) datasource.DataSource {
	return &dataController[T]{
		name:         name,
		view:         view,
		newSchema:    newSchema,
		newDataModel: newDataModel,
	}
}

type dataController[T any] struct {
	// name is the name of the resource or datasource,
	// for instance, "aiven_organization_address"
	name string

	// view implements CRUD functions
	view View[T]

	// newSchema returns a new instance of the generated schema.
	newSchema datasourceSchema

	// newDataModel returns a new instance of the generated dataModel.
	newDataModel dataModelFactory[T]
}

func (c *dataController[T]) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	rsp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		// TF calls Configure many times, it might not contain the provider data yet.
		return
	}

	p, ok := req.ProviderData.(types.AivenClientProvider)
	if !ok {
		rsp.Diagnostics.AddError(
			errmsg.SummaryUnexpectedProviderDataType,
			fmt.Sprintf(errmsg.DetailUnexpectedProviderDataType, req.ProviderData),
		)
		return
	}

	c.view.Configure(p.GetGenClient())
}

func (c *dataController[T]) Metadata(
	_ context.Context,
	_ datasource.MetadataRequest,
	rsp *datasource.MetadataResponse,
) {
	rsp.TypeName = c.name
}

func (c *dataController[T]) Schema(
	ctx context.Context,
	_ datasource.SchemaRequest,
	rsp *datasource.SchemaResponse,
) {
	rsp.Schema = c.newSchema(ctx)
}

func (c *dataController[T]) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	rsp *datasource.ReadResponse,
) {
	var (
		state = c.newDataModel()
		diags = &rsp.Diagnostics
	)
	diags.Append(req.Config.Get(ctx, state)...)
	if diags.HasError() {
		return
	}

	diags.Append(c.view.Read(ctx, state.DataModel())...)
	if diags.HasError() {
		return
	}

	diags.Append(rsp.State.Set(ctx, state)...)
}
