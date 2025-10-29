package adapter

import (
	"context"
	"fmt"
	"reflect"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/providerdata"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// Model implements resource or datasource model with the shared fields model.
type Model[T any] interface {
	// SharedModel returns the shared fields model between resource and datasource.
	SharedModel() *T
	TimeoutsObject() types.Object
}

// instantiate creates a new instance of the model using reflection since we can't
// directly instantiate an interface type.
func instantiate[M Model[T], T any]() M {
	var m M
	return reflect.New(reflect.TypeOf(m).Elem()).Interface().(M)
}

type ResourceOptions[M Model[T], T any] struct {
	// TypeName is the name of resource,
	// for instance, "aiven_organization_address"
	TypeName string
	Schema   func(ctx context.Context) schema.Schema

	// IDFields are used to build the resource ID.
	// Example: ["project_id", "instance_name"] == "project-123/instance-456"
	IDFields []string

	// Whether to call Read after Create and Update operations.
	RefreshState bool

	// CRUD operations
	Read   func(ctx context.Context, client avngen.Client, state *T) diag.Diagnostics
	Delete func(ctx context.Context, client avngen.Client, state *T) diag.Diagnostics
	Create func(ctx context.Context, client avngen.Client, plan *T) diag.Diagnostics
	Update func(ctx context.Context, client avngen.Client, plan, state, config *T) diag.Diagnostics

	// ModifyPlan implements resource.ResourceWithModifyPlan.
	// https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#resource-plan-modification
	ModifyPlan func(ctx context.Context, client avngen.Client, plan, state, config *T) diag.Diagnostics

	// ValidateConfig implements resource.ResourceWithValidateConfig.
	// https://developer.hashicorp.com/terraform/plugin/framework/resources/validate-configuration#validateconfig-method
	ValidateConfig func(ctx context.Context, client avngen.Client, config *T) diag.Diagnostics

	// ConfigValidators implements resource.ResourceWithConfigValidators.
	// https://developer.hashicorp.com/terraform/plugin/framework/resources/validate-configuration#configvalidators-method
	ConfigValidators func(ctx context.Context, client avngen.Client) []resource.ConfigValidator
}

func NewResource[M Model[T], T any](options ResourceOptions[M, T]) resource.Resource {
	return &resourceAdapter[M, T]{
		resource: options,
	}
}

// NewLazyResource creates a lazy resource constructor.
// The provider.Provider.Resources requires a function that returns a resource.Resource.
func NewLazyResource[M Model[T], T any](options ResourceOptions[M, T]) func() resource.Resource {
	return func() resource.Resource {
		return NewResource(options)
	}
}

var (
	_ resource.Resource                     = (*resourceAdapter[Model[any], any])(nil)
	_ resource.ResourceWithConfigure        = (*resourceAdapter[Model[any], any])(nil)
	_ resource.ResourceWithImportState      = (*resourceAdapter[Model[any], any])(nil)
	_ resource.ResourceWithValidateConfig   = (*resourceAdapter[Model[any], any])(nil)
	_ resource.ResourceWithConfigValidators = (*resourceAdapter[Model[any], any])(nil)
	_ resource.ResourceWithModifyPlan       = (*resourceAdapter[Model[any], any])(nil)
)

type resourceAdapter[M Model[T], T any] struct {
	client   avngen.Client
	resource ResourceOptions[M, T]
}

func (a *resourceAdapter[M, T]) Configure(
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

	a.client = p.GetGenClient()
}

func (a *resourceAdapter[M, T]) Metadata(
	_ context.Context,
	_ resource.MetadataRequest,
	rsp *resource.MetadataResponse,
) {
	rsp.TypeName = a.resource.TypeName
}

func (a *resourceAdapter[M, T]) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	rsp *resource.SchemaResponse,
) {
	rsp.Schema = a.resource.Schema(ctx)
}

func (a *resourceAdapter[M, T]) Create(
	ctx context.Context,
	req resource.CreateRequest,
	rsp *resource.CreateResponse,
) {
	var (
		plan  = instantiate[M]()
		diags = &rsp.Diagnostics
	)

	diags.Append(req.Plan.Get(ctx, plan)...)
	if diags.HasError() {
		return
	}

	ctx, cancel, d := withTimeout(ctx, plan.TimeoutsObject(), timeoutCreate)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	defer cancel()

	diags.Append(a.resource.Create(ctx, a.client, plan.SharedModel())...)
	if diags.HasError() {
		return
	}

	if a.resource.RefreshState {
		diags.Append(a.refreshState(ctx, plan)...)
		if diags.HasError() {
			return
		}
	}

	diags.Append(rsp.State.Set(ctx, plan)...)
}

func (a *resourceAdapter[M, T]) Read(
	ctx context.Context,
	req resource.ReadRequest,
	rsp *resource.ReadResponse,
) {
	var (
		state = instantiate[M]()
		diags = &rsp.Diagnostics
	)

	diags.Append(req.State.Get(ctx, state)...)
	if diags.HasError() {
		return
	}

	ctx, cancel, d := withTimeout(ctx, state.TimeoutsObject(), timeoutRead)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	defer cancel()

	diags.Append(a.resource.Read(ctx, a.client, state.SharedModel())...)
	if diags.HasError() {
		return
	}

	diags.Append(rsp.State.Set(ctx, state)...)
}

// refreshState retries Read if the resource is not found.
// In rare cases, the backend might return 404 after the resource is created or updated.
func (a *resourceAdapter[M, T]) refreshState(ctx context.Context, plan M) diag.Diagnostics {
	return errmsg.RetryDiags(
		func() diag.Diagnostics {
			return a.resource.Read(ctx, a.client, plan.SharedModel())
		},
		retry.Context(ctx),
		retry.RetryIf(avngen.IsNotFound),
	)
}

func (a *resourceAdapter[M, T]) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	rsp *resource.UpdateResponse,
) {
	if a.resource.Update == nil {
		return
	}

	var (
		plan   = instantiate[M]()
		state  = instantiate[M]()
		config = instantiate[M]()
		diags  = &rsp.Diagnostics
	)

	diags.Append(req.Plan.Get(ctx, plan)...)
	diags.Append(req.State.Get(ctx, state)...)
	diags.Append(req.Config.Get(ctx, config)...)
	if diags.HasError() {
		return
	}

	ctx, cancel, d := withTimeout(ctx, plan.TimeoutsObject(), timeoutUpdate)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	defer cancel()

	diags.Append(a.resource.Update(ctx, a.client, plan.SharedModel(), state.SharedModel(), config.SharedModel())...)
	if diags.HasError() {
		return
	}

	if a.resource.RefreshState {
		diags.Append(a.refreshState(ctx, plan)...)
		if diags.HasError() {
			return
		}
	}

	diags.Append(rsp.State.Set(ctx, plan)...)
}

func (a *resourceAdapter[M, T]) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	rsp *resource.DeleteResponse,
) {
	var (
		state = instantiate[M]()
		diags = &rsp.Diagnostics
	)

	diags.Append(req.State.Get(ctx, state)...)
	if diags.HasError() {
		return
	}

	ctx, cancel, d := withTimeout(ctx, state.TimeoutsObject(), timeoutDelete)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	defer cancel()

	diags.Append(a.resource.Delete(ctx, a.client, state.SharedModel())...)
}

func (a *resourceAdapter[M, T]) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	rsp *resource.ImportStateResponse,
) {
	values, err := schemautil.SplitResourceID(req.ID, len(a.resource.IDFields))
	if err != nil {
		importPath := schemautil.BuildResourceID(a.resource.IDFields...)
		rsp.Diagnostics.AddError(
			"Unexpected Read Identifier",
			fmt.Sprintf("Expected import identifier with format: %q. Got: %q", importPath, req.ID),
		)
		return
	}

	for i, v := range values {
		rsp.Diagnostics.Append(rsp.State.SetAttribute(ctx, path.Root(a.resource.IDFields[i]), v)...)
	}
}

func (a *resourceAdapter[M, T]) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	if a.resource.ConfigValidators == nil {
		return nil
	}
	return a.resource.ConfigValidators(ctx, a.client)
}

func (a *resourceAdapter[M, T]) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	rsp *resource.ValidateConfigResponse,
) {
	if a.resource.ValidateConfig == nil {
		return
	}

	var (
		config = instantiate[M]()
		diags  = &rsp.Diagnostics
	)
	diags.Append(req.Config.Get(ctx, config)...)
	if diags.HasError() {
		return
	}

	// Some resources might run API calls to validate the configuration.
	ctx, cancel, d := withTimeout(ctx, config.TimeoutsObject(), timeoutRead)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	defer cancel()

	diags.Append(a.resource.ValidateConfig(ctx, a.client, config.SharedModel())...)
}

func (a *resourceAdapter[M, T]) ModifyPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	rsp *resource.ModifyPlanResponse,
) {
	if a.resource.ModifyPlan == nil {
		return
	}

	var (
		plan   = instantiate[M]()
		state  = instantiate[M]()
		config = instantiate[M]()
		diags  = &rsp.Diagnostics
	)

	if !req.Plan.Raw.IsNull() {
		// If resource is not marked for deletion.
		diags.Append(req.Plan.Get(ctx, plan)...)
		diags.Append(req.Config.Get(ctx, config)...)
	}
	diags.Append(req.State.Get(ctx, state)...)
	if diags.HasError() {
		return
	}

	ctx, cancel, d := withTimeout(ctx, plan.TimeoutsObject(), timeoutRead)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	defer cancel()

	diags.Append(a.resource.ModifyPlan(ctx, a.client, plan.SharedModel(), state.SharedModel(), config.SharedModel())...)
	if diags.HasError() {
		return
	}

	diags.Append(rsp.Plan.Set(ctx, plan)...)
}
