package adapter

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/providerdata"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// MightyResource implements additional resource methods
type MightyResource interface {
	resource.ResourceWithConfigure
	resource.ResourceWithImportState
}

type resourceSchema func(context.Context) schema.Schema

func NewResource[T any](
	name string,
	view ResView[T],
	newSchema resourceSchema,
	newDataModel dataModelFactory[T],
	idFields []string,
) MightyResource {
	return &resourceAdapter[T]{
		name:         name,
		view:         view,
		newSchema:    newSchema,
		newDataModel: newDataModel,
		idFields:     idFields,
	}
}

type resourceAdapter[T any] struct {
	name         string
	view         ResView[T]
	newSchema    resourceSchema
	newDataModel dataModelFactory[T]
	idFields     []string
}

func (a *resourceAdapter[T]) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	rsp *resource.ConfigureResponse,
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

	// Setups the client.
	// Configure other things, like logging, if needed.
	a.view.Configure(p.GetGenClient())
}

func (a *resourceAdapter[T]) Metadata(
	_ context.Context,
	_ resource.MetadataRequest,
	rsp *resource.MetadataResponse,
) {
	rsp.TypeName = a.name
}

func (a *resourceAdapter[T]) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	rsp *resource.SchemaResponse,
) {
	rsp.Schema = a.newSchema(ctx)
}

func (a *resourceAdapter[T]) Create(
	ctx context.Context,
	req resource.CreateRequest,
	rsp *resource.CreateResponse,
) {
	var (
		plan  = a.newDataModel()
		diags = &rsp.Diagnostics
	)

	diags.Append(req.Plan.Get(ctx, plan)...)
	if diags.HasError() {
		return
	}

	diags.Append(a.view.Create(ctx, plan.DataModel())...)
	if diags.HasError() {
		return
	}

	diags.Append(rsp.State.Set(ctx, plan)...)
}

func (a *resourceAdapter[T]) Read(
	ctx context.Context,
	req resource.ReadRequest,
	rsp *resource.ReadResponse,
) {
	var (
		state = a.newDataModel()
		diags = &rsp.Diagnostics
	)
	diags.Append(req.State.Get(ctx, state)...)
	if diags.HasError() {
		return
	}

	diags.Append(a.view.Read(ctx, state.DataModel())...)
	if diags.HasError() {
		return
	}

	diags.Append(rsp.State.Set(ctx, state)...)
}

func (a *resourceAdapter[T]) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	rsp *resource.UpdateResponse,
) {
	var (
		plan  = a.newDataModel()
		state = a.newDataModel()
		diags = &rsp.Diagnostics
	)

	diags.Append(req.Plan.Get(ctx, plan)...)
	diags.Append(req.State.Get(ctx, state)...)
	if diags.HasError() {
		return
	}

	diags.Append(a.view.Update(ctx, plan.DataModel(), state.DataModel())...)
	if diags.HasError() {
		return
	}

	diags.Append(rsp.State.Set(ctx, plan)...)
}

func (a *resourceAdapter[T]) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	rsp *resource.DeleteResponse,
) {
	var (
		state = a.newDataModel()
		diags = &rsp.Diagnostics
	)
	diags.Append(req.State.Get(ctx, state)...)
	if diags.HasError() {
		return
	}

	diags.Append(a.view.Delete(ctx, state.DataModel())...)
}

func (a *resourceAdapter[T]) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	rsp *resource.ImportStateResponse,
) {
	values, err := schemautil.SplitResourceID(req.ID, len(a.idFields))
	if err != nil {
		importPath := schemautil.BuildResourceID(a.idFields...)
		rsp.Diagnostics.AddError(
			"Unexpected Read Identifier",
			fmt.Sprintf("Expected import identifier with format: %q. Got: %q", importPath, req.ID),
		)
	}

	for i, v := range values {
		rsp.Diagnostics.Append(rsp.State.SetAttribute(ctx, path.Root(a.idFields[i]), v)...)
	}
}
