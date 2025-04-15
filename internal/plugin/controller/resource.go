package controller

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/types"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// NewResource returns resource.Resource but must support Configure and ImportState
var _ resource.ResourceWithConfigure = (*resController[any])(nil)
var _ resource.ResourceWithImportState = (*resController[any])(nil)

type resourceSchema func(context.Context) schema.Schema

func NewResource[T any](
	name string,
	view View[T],
	newSchema resourceSchema,
	newDataModel dataModelFactory[T],
	idFields []string,
) resource.Resource {
	return &resController[T]{
		name:         name,
		view:         view,
		newSchema:    newSchema,
		newDataModel: newDataModel,
		idFields:     idFields,
	}
}

type resController[T any] struct {
	name         string
	view         View[T]
	newSchema    resourceSchema
	newDataModel dataModelFactory[T]
	idFields     []string
}

func (c *resController[T]) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	rsp *resource.ConfigureResponse,
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

	// Setups the client.
	// Configure other things, like logging, if needed.
	c.view.Configure(p.GetGenClient())
}

func (c *resController[T]) Metadata(
	_ context.Context,
	_ resource.MetadataRequest,
	rsp *resource.MetadataResponse,
) {
	rsp.TypeName = c.name
}

func (c *resController[T]) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	rsp *resource.SchemaResponse,
) {
	rsp.Schema = c.newSchema(ctx)
}

func (c *resController[T]) Create(
	ctx context.Context,
	req resource.CreateRequest,
	rsp *resource.CreateResponse,
) {
	var (
		plan  = c.newDataModel()
		diags = &rsp.Diagnostics
	)

	diags.Append(req.Plan.Get(ctx, plan)...)
	if diags.HasError() {
		return
	}

	diags.Append(c.view.Create(ctx, plan.DataModel())...)
	if diags.HasError() {
		return
	}

	diags.Append(rsp.State.Set(ctx, plan)...)
}

func (c *resController[T]) Read(
	ctx context.Context,
	req resource.ReadRequest,
	rsp *resource.ReadResponse,
) {
	var (
		state = c.newDataModel()
		diags = &rsp.Diagnostics
	)
	diags.Append(req.State.Get(ctx, state)...)
	if diags.HasError() {
		return
	}

	diags.Append(c.view.Read(ctx, state.DataModel())...)
	if diags.HasError() {
		return
	}

	diags.Append(rsp.State.Set(ctx, state)...)
}

func (c *resController[T]) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	rsp *resource.UpdateResponse,
) {
	var (
		plan  = c.newDataModel()
		state = c.newDataModel()
		diags = &rsp.Diagnostics
	)

	diags.Append(req.Plan.Get(ctx, plan)...)
	diags.Append(req.State.Get(ctx, state)...)
	if diags.HasError() {
		return
	}

	diags.Append(c.view.Update(ctx, plan.DataModel(), state.DataModel())...)
	if diags.HasError() {
		return
	}

	diags.Append(rsp.State.Set(ctx, plan)...)
}

func (c *resController[T]) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	rsp *resource.DeleteResponse,
) {
	var (
		state = c.newDataModel()
		diags = &rsp.Diagnostics
	)
	diags.Append(req.State.Get(ctx, state)...)
	if diags.HasError() {
		return
	}

	diags.Append(c.view.Delete(ctx, state.DataModel())...)
}

func (c *resController[T]) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	rsp *resource.ImportStateResponse,
) {
	values, err := schemautil.SplitResourceID(req.ID, len(c.idFields))
	if err != nil {
		importPath := schemautil.BuildResourceID(c.idFields...)
		rsp.Diagnostics.AddError(
			"Unexpected Read Identifier",
			fmt.Sprintf("Expected import identifier with format: %q. Got: %q", importPath, req.ID),
		)
	}

	for i, v := range values {
		rsp.Diagnostics.Append(rsp.State.SetAttribute(ctx, path.Root(c.idFields[i]), v)...)
	}
}
