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
	resource.ResourceWithValidateConfig
	resource.ResourceWithConfigValidators
}

type newResourceSchema func(context.Context) schema.Schema

func NewResource[T any](
	name string,
	view ResView[T],
	newSchema newResourceSchema,
	newModel newModel[T],
	composeID []string,
) MightyResource {
	return &resourceAdapter[T]{
		name:      name,
		view:      view,
		newSchema: newSchema,
		newModel:  newModel,
		composeID: composeID,
	}
}

type resourceAdapter[T any] struct {
	// name is the name of resource,
	// for instance, "aiven_organization_address"
	name string

	// view implements CRUD functions
	view ResView[T]

	// newSchema returns a new instance of the generated resource Schema.
	newSchema newResourceSchema

	// newModel returns a new instance of the generated datasource newModel.
	newModel newModel[T]

	// composeID is the list of identifiers used to compose the resource ID.
	composeID []string
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
		plan  = a.newModel()
		diags = &rsp.Diagnostics
	)

	diags.Append(req.Plan.Get(ctx, plan)...)
	if diags.HasError() {
		return
	}

	diags.Append(a.view.Create(ctx, plan.SharedModel())...)
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
		state = a.newModel()
		diags = &rsp.Diagnostics
	)
	diags.Append(req.State.Get(ctx, state)...)
	if diags.HasError() {
		return
	}

	diags.Append(a.view.Read(ctx, state.SharedModel())...)
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
		plan  = a.newModel()
		state = a.newModel()
		diags = &rsp.Diagnostics
	)

	diags.Append(req.Plan.Get(ctx, plan)...)
	diags.Append(req.State.Get(ctx, state)...)
	if diags.HasError() {
		return
	}

	diags.Append(a.view.Update(ctx, plan.SharedModel(), state.SharedModel())...)
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
		state = a.newModel()
		diags = &rsp.Diagnostics
	)
	diags.Append(req.State.Get(ctx, state)...)
	if diags.HasError() {
		return
	}

	diags.Append(a.view.Delete(ctx, state.SharedModel())...)
}

func (a *resourceAdapter[T]) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	rsp *resource.ImportStateResponse,
) {
	values, err := schemautil.SplitResourceID(req.ID, len(a.composeID))
	if err != nil {
		importPath := schemautil.BuildResourceID(a.composeID...)
		rsp.Diagnostics.AddError(
			"Unexpected Read Identifier",
			fmt.Sprintf("Expected import identifier with format: %q. Got: %q", importPath, req.ID),
		)
	}

	for i, v := range values {
		rsp.Diagnostics.Append(rsp.State.SetAttribute(ctx, path.Root(a.composeID[i]), v)...)
	}
}

func (a *resourceAdapter[T]) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	v, ok := a.view.(ResConfigValidators[T])
	if !ok {
		return nil
	}
	return v.ResConfigValidators(ctx)
}

func (a *resourceAdapter[T]) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	rsp *resource.ValidateConfigResponse,
) {
	v, ok := a.view.(ResValidateConfig[T])
	if !ok {
		return
	}

	var (
		config = a.newModel()
		diags  = &rsp.Diagnostics
	)
	diags.Append(req.Config.Get(ctx, config)...)
	if diags.HasError() {
		return
	}

	diags.Append(v.ResValidateConfig(ctx, config.SharedModel())...)
}
