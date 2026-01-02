package adapter

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/providerdata"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
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

	// Indicates whether the resource is in beta.
	// Requires the `PROVIDER_AIVEN_ENABLE_BETA` environment variable to be set.
	Beta bool

	// IDFields are used to build the resource ID.
	// Example: ["project_id", "instance_name"] == "project-123/instance-456"
	IDFields []string

	// Whether to call Read after Create and Update operations.
	// Retries common errors like 404 and 403 (see the implementation for details).
	RefreshState bool

	// RemoveMissing removes the resource from the state if it's missing (i.e., if Read() returns an avngen.IsNotFound error).
	// This is useful for resources that Aiven may automatically delete,
	// such as users or databases when a service has been turned off and then on again.
	// Instead of forcing users to manually clean up the state, Terraform will plan to "create" the resource again.
	RemoveMissing bool

	// Throws an error if the resource is marked for deletion and `termination_protection` field is set to true.
	// "Virtual" fields (not managed by the API) should be deprecated. Instead, use "prevent_destroy":
	// https://developer.hashicorp.com/terraform/tutorials/state/resource-lifecycle#prevent-resource-deletion
	TerminationProtection bool

	// CRUD operations.
	// Each CRUD operation should return diag.Diagnostics (rather than taking it as an argument)
	// This design allows internal retry logic for operations that can fail transiently.
	// NOTE: Create and Update must NOT invoke Read themselves; set RefreshState=true to trigger a post-operation Read.
	// NOTE2: Delete ignores 404 errors since resources may already be deleted after client retries.
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

	// When RemoveResource is enabled, we remove the resource from the state if it's missing.
	// See ResourceOptions.RemoveMissing for more details.
	readDiags := a.resource.Read(ctx, a.client, state.SharedModel())
	if a.resource.RemoveMissing && errmsg.HasDiagError(readDiags, avngen.IsNotFound) {
		// Ignores all the other diagnostics and removes the resource from the state.
		rsp.State.RemoveResource(ctx)
		return
	}

	diags.Append(readDiags...)
	if diags.HasError() {
		return
	}

	diags.Append(rsp.State.Set(ctx, state)...)
}

// refreshState reads the resource state after Create or Update operation.
// In rare cases, the backend might be inconsistent right after yet, retries known errors.
func (a *resourceAdapter[M, T]) refreshState(ctx context.Context, plan M) diag.Diagnostics {
	return errmsg.RetryDiags(
		ctx,
		func() diag.Diagnostics {
			return a.resource.Read(ctx, a.client, plan.SharedModel())
		},
		retry.Delay(time.Second*5),
		retry.Attempts(10),
		retry.LastErrorOnly(true),
		errmsg.RetryIfAivenStatus(
			http.StatusNotFound,  // The API is inconsistent, returns 404 after Create/Update
			http.StatusForbidden, // Eventual consistency might cause permission errors, for instance for "organization_project"
		),
	)
}

func (a *resourceAdapter[M, T]) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	rsp *resource.UpdateResponse,
) {
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

	// Some resources might have "virtual" fields, like "termination_protection".
	// Those fields can be technically updated, but they don't require an API call.
	// So we skip the Update call if it's not implemented.
	if a.resource.Update != nil {
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

	// The Aiven client might receive 5xx errors from the backend, causing it to retry the delete operation.
	// However, the resource may have already been deleted, in which case a 404 error can be safely ignored.
	d = a.resource.Delete(ctx, a.client, state.SharedModel())
	diags.Append(errmsg.DropDiagError(d, avngen.IsNotFound)...)
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
	if a.resource.Beta && !util.IsBeta() {
		rsp.Diagnostics.AddError(
			"Beta Resource Not Enabled",
			fmt.Sprintf("The `%s` resource is in beta. Set the `%s` environment variable to enable.", a.resource.TypeName, util.AivenEnableBeta),
		)
		return
	}

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
	// Checks if termination protection is enabled for this resource
	if a.resource.TerminationProtection && req.Plan.Raw.IsNull() {
		// req.Plan.Raw.IsNull() means "marked for deletion":
		// https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#resource-destroy-plan-diagnostics
		enabled := false
		rsp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("termination_protection"), &enabled)...)
		if rsp.Diagnostics.HasError() {
			return
		}

		if enabled {
			rsp.Diagnostics.AddError(
				errmsg.SummaryErrorDeletingResource,
				fmt.Sprintf("The resource `%s` has termination protection enabled and cannot be deleted.", a.resource.TypeName),
			)
			return
		}
	}

	if a.resource.ModifyPlan == nil {
		return
	}

	if req.Plan.Raw.IsNull() {
		// There is no plan to modify if the resource is marked for deletion
		return
	}

	var (
		plan   = instantiate[M]()
		config = instantiate[M]()
		diags  = &rsp.Diagnostics
	)

	diags.Append(req.Plan.Get(ctx, plan)...)
	diags.Append(req.Config.Get(ctx, config)...)
	if diags.HasError() {
		return
	}

	var sharedState *T
	if !req.State.Raw.IsNull() {
		// During create planning, state is null. Let the ModifyPlan hook decide how to handle a nil state
		state := instantiate[M]()
		diags.Append(req.State.Get(ctx, state)...)
		if diags.HasError() {
			return
		}
		sharedState = state.SharedModel()
	}

	ctx, cancel, d := withTimeout(ctx, plan.TimeoutsObject(), timeoutRead)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	defer cancel()

	diags.Append(a.resource.ModifyPlan(ctx, a.client, plan.SharedModel(), sharedState, config.SharedModel())...)
	if diags.HasError() {
		return
	}

	diags.Append(rsp.Plan.Set(ctx, plan)...)
}
