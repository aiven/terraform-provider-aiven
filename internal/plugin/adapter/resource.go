package adapter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/providerdata"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

type ResourceOptions struct {
	// TypeName is the name of resource,
	// for instance, "aiven_organization_address"
	TypeName       string
	Schema         func(ctx context.Context) schema.Schema
	SchemaInternal *Schema

	// Indicates whether the resource is in beta.
	// Requires the `PROVIDER_AIVEN_ENABLE_BETA` environment variable to be set.
	Beta bool

	// IDFields are used to build the resource ID.
	// Example: ["project_id", "instance_name"] == "project-123/instance-456"
	IDFields []string

	// Whether to call Read after Create and Update operations.
	// Retries common errors like 404 and 403 (see the implementation for details).
	RefreshState bool

	// Time to wait after creating or updating the resource to let the backend catch up.
	RefreshStateDelay time.Duration

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

	Create func(ctx context.Context, client avngen.Client, d ResourceData) error
	Update func(ctx context.Context, client avngen.Client, d ResourceData) error
	Delete func(ctx context.Context, client avngen.Client, d ResourceData) error
	Read   func(ctx context.Context, client avngen.Client, d ResourceData) error

	// ModifyPlan implements resource.ResourceWithModifyPlan.
	// https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#resource-plan-modification
	ModifyPlan func(ctx context.Context, client avngen.Client, d ResourceData) error

	// ValidateConfig implements resource.ResourceWithValidateConfig.
	// https://developer.hashicorp.com/terraform/plugin/framework/resources/validate-configuration#validateconfig-method
	ValidateConfig func(ctx context.Context, client avngen.Client, d ResourceData) error

	// ConfigValidators implements resource.ResourceWithConfigValidators.
	// https://developer.hashicorp.com/terraform/plugin/framework/resources/validate-configuration#configvalidators-method
	ConfigValidators func(ctx context.Context, client avngen.Client) []resource.ConfigValidator
}

func NewResource(options ResourceOptions) resource.Resource {
	return &resourceAdapter{
		resource: options,
	}
}

// NewLazyResource creates a lazy resource constructor.
// The provider.Provider.Resources requires a function that returns a resource.Resource.
func NewLazyResource(options ResourceOptions) func() resource.Resource {
	return func() resource.Resource {
		return NewResource(options)
	}
}

var (
	_ resource.Resource                     = (*resourceAdapter)(nil)
	_ resource.ResourceWithConfigure        = (*resourceAdapter)(nil)
	_ resource.ResourceWithImportState      = (*resourceAdapter)(nil)
	_ resource.ResourceWithValidateConfig   = (*resourceAdapter)(nil)
	_ resource.ResourceWithConfigValidators = (*resourceAdapter)(nil)
	_ resource.ResourceWithModifyPlan       = (*resourceAdapter)(nil)
)

type resourceAdapter struct {
	client   avngen.Client
	resource ResourceOptions
}

func (a *resourceAdapter) Configure(
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

func (a *resourceAdapter) Metadata(
	_ context.Context,
	_ resource.MetadataRequest,
	rsp *resource.MetadataResponse,
) {
	rsp.TypeName = a.resource.TypeName
}

func (a *resourceAdapter) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	rsp *resource.SchemaResponse,
) {
	rsp.Schema = a.resource.Schema(ctx)
}

func (a *resourceAdapter) Create(
	ctx context.Context,
	req resource.CreateRequest,
	rsp *resource.CreateResponse,
) {
	diags := &rsp.Diagnostics

	d, err := NewResourceData(a.resource.SchemaInternal, a.resource.IDFields, &req.Plan, nil, &req.Config)
	if err != nil {
		diags.AddError("failed to create ResourceData", err.Error())
		return
	}

	ctx, cancel, err := withTimeout(ctx, d, timeoutCreate)
	if err != nil {
		diags.AddError("failed to read timeouts block", err.Error())
		return
	}
	defer cancel()

	err = a.resource.Create(ctx, a.client, d)
	if err != nil {
		diags.AddError("failed to create resource", err.Error())
		return
	}

	if a.resource.RefreshState {
		err = a.refreshState(ctx, d)
		if err != nil {
			diags.AddError("failed to refresh state", err.Error())
			return
		}
	}

	rsp.State.Raw = d.tfValue()
}

func (a *resourceAdapter) Read(
	ctx context.Context,
	req resource.ReadRequest,
	rsp *resource.ReadResponse,
) {
	diags := &rsp.Diagnostics

	d, err := NewResourceData(a.resource.SchemaInternal, a.resource.IDFields, nil, &req.State, nil)
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

	// When RemoveResource is enabled, we remove the resource from the state if it's missing.
	// See ResourceOptions.RemoveMissing for more details.
	err = a.resource.Read(ctx, a.client, d)
	if a.resource.RemoveMissing && avngen.IsNotFound(err) {
		// Ignores all the other diagnostics and removes the resource from the state.
		rsp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		diags.AddError("failed to read resource", err.Error())
		return
	}

	rsp.State.Raw = d.tfValue()
}

// refreshState reads the resource state after Create or Update operation.
// Optionally waits RefreshStateDelay before reading. In rare cases, the backend might be
// inconsistent right after the operation; retries known errors.
func (a *resourceAdapter) refreshState(ctx context.Context, rd ResourceData) error {
	if a.resource.RefreshStateDelay != 0 {
		delay := time.NewTimer(a.resource.RefreshStateDelay)
		defer delay.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-delay.C:
		}
	}
	return retry.Do(
		func() error {
			return a.resource.Read(ctx, a.client, rd)
		},
		retry.Delay(time.Second*5),
		retry.Attempts(10),
		retry.LastErrorOnly(true),
		retry.RetryIf(func(err error) bool {
			var e avngen.Error
			// 404: The API is inconsistent, returns 404 after Create/Update
			// 403: Eventual consistency might cause permission errors, for instance for "organization_project"
			return errors.As(err, &e) && (e.Status == http.StatusNotFound || e.Status == http.StatusForbidden)
		}),
	)
}

func (a *resourceAdapter) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	rsp *resource.UpdateResponse,
) {
	diags := &rsp.Diagnostics

	d, err := NewResourceData(a.resource.SchemaInternal, a.resource.IDFields, &req.Plan, &req.State, &req.Config)
	if err != nil {
		diags.AddError("failed to create ResourceData", err.Error())
		return
	}

	ctx, cancel, err := withTimeout(ctx, d, timeoutUpdate)
	if err != nil {
		diags.AddError("failed to read timeouts block", err.Error())
		return
	}
	defer cancel()

	// Some resources might have "virtual" fields, like "termination_protection".
	// Those fields can be technically updated, but they don't require an API call.
	// So we skip the Update call if it's not implemented.
	if a.resource.Update != nil {
		err = a.resource.Update(ctx, a.client, d)
		if err != nil {
			diags.AddError("failed to update resource", err.Error())
			return
		}
	}

	if a.resource.RefreshState {
		err = a.refreshState(ctx, d)
		if err != nil {
			diags.AddError("failed to refresh state", err.Error())
			return
		}
	}

	rsp.State.Raw = d.tfValue()
}

func (a *resourceAdapter) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	rsp *resource.DeleteResponse,
) {
	diags := &rsp.Diagnostics

	d, err := NewResourceData(a.resource.SchemaInternal, a.resource.IDFields, nil, &req.State, nil)
	if err != nil {
		diags.AddError("failed to create ResourceData", err.Error())
		return
	}

	ctx, cancel, err := withTimeout(ctx, d, timeoutDelete)
	if err != nil {
		diags.AddError("failed to read timeouts block", err.Error())
		return
	}
	defer cancel()

	// The Aiven client might receive 5xx errors from the backend, causing it to retry the delete operation.
	// However, the resource may have already been deleted, in which case a 404 error can be safely ignored.
	err = a.resource.Delete(ctx, a.client, d)
	if err != nil && !avngen.IsNotFound(err) {
		diags.AddError("failed to delete resource", err.Error())
	}
}

func (a *resourceAdapter) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	rsp *resource.ImportStateResponse,
) {
	// Set only ID fields here; Terraform runs Read afterward to populate full state.
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

func (a *resourceAdapter) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	if a.resource.ConfigValidators == nil {
		return nil
	}
	return a.resource.ConfigValidators(ctx, a.client)
}

func (a *resourceAdapter) ValidateConfig(
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

	diags := &rsp.Diagnostics

	d, err := NewResourceData(a.resource.SchemaInternal, a.resource.IDFields, nil, nil, &req.Config)
	if err != nil {
		diags.AddError("failed to create ResourceData", err.Error())
		return
	}

	// Some resources might run API calls to validate the configuration.
	ctx, cancel, err := withTimeout(ctx, d, timeoutRead)
	if err != nil {
		diags.AddError("failed to read timeouts block", err.Error())
		return
	}
	defer cancel()

	err = a.resource.ValidateConfig(ctx, a.client, d)
	if err != nil {
		diags.AddError("failed to validate config", err.Error())
		return
	}
}

func (a *resourceAdapter) ModifyPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	rsp *resource.ModifyPlanResponse,
) {
	if req.Plan.Raw.IsNull() {
		// Checks if termination protection is enabled for this resource
		if a.resource.TerminationProtection {
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
			}
		}

		// There is no plan to modify if the resource is marked for deletion
		return
	}

	if a.resource.ModifyPlan == nil {
		return
	}

	diags := &rsp.Diagnostics
	var stateOrNil *tfsdk.State
	if !req.State.Raw.IsNull() {
		stateOrNil = &req.State
	}

	d, err := NewResourceData(a.resource.SchemaInternal, a.resource.IDFields, &req.Plan, stateOrNil, &req.Config)
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

	err = a.resource.ModifyPlan(ctx, a.client, d)
	if err != nil {
		diags.AddError("failed to modify plan", err.Error())
		return
	}

	rsp.Plan.Raw = d.tfValue()
}

func MustParseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		panic(fmt.Errorf("failed to parse duration: %w", err))
	}
	return d
}
