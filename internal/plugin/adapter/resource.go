package adapter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/avast/retry-go/v4"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"

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
	// Implemented by refreshState; see that function for retry behavior.
	RefreshState bool

	// Time to wait after creating or updating the resource to let the backend catch up.
	RefreshStateDelay time.Duration

	// Attribute names and values that must match after Read before refreshState succeeds.
	// Mismatches return ErrRefreshStateDesired and are retried with the same backoff as transient Read errors.
	RefreshStateDesired map[string]string

	// refreshStateRetryDelay is the base delay used by retry-go's BackOff between Read attempts in refreshState.
	// Zero falls back to the package default (defaultRefreshStateRetryDelay). Unexported: intended for tests
	// inside the adapter package only; not wired through the YAML schema or the generator.
	refreshStateRetryDelay time.Duration

	// DeleteState, when non-nil, enables the delete poller: after issuing Delete, the adapter polls
	// Read until the resource reaches its terminal state. See DeleteStateOptions and deleteState.
	DeleteState *DeleteStateOptions

	// RemoveMissing removes the resource from the state if it's missing (i.e., if Read() returns an IsNotFound error).
	// This is useful for resources that Aiven may automatically delete,
	// such as users or databases when a service has been turned off and then on again.
	// Instead of forcing users to manually clean up the state, Terraform will plan to "create" the resource again.
	RemoveMissing bool

	// IgnoreAlreadyExists treats a 409 Conflict ("already exists") response from Create as success.
	// This is useful when the resource being managed is a one-shot provisioning action whose effect is
	// idempotent from Terraform's perspective — re-running it should adopt the existing entity rather
	// than fail. Requires RefreshState=true so that the full state is populated by the subsequent Read.
	// NOTE: This does NOT work for resources whose ID is returned only in the Create response
	// (e.g. server-assigned IDs), because on conflict there is no response to read the ID from and
	// the subsequent Read has nothing to look up.
	IgnoreAlreadyExists bool

	// Returns an error on "plan" if the resource is marked for deletion and the `termination_protection` field is set to true.
	// There are two types of resources with a `termination_protection` field in our provider:
	// 1. API-level termination protection: defined in the API schema and prevents deletion through the API.
	// 2. Client-side termination protection: defined only in Terraform and prevents deletion in Terraform,
	//    but the resource can still be deleted via the Aiven Console, CLI, etc.
	//    These "virtual" fields should be deprecated and removed in the future.
	//    Instead, use the recommended "prevent_destroy":
	//    https://developer.hashicorp.com/terraform/tutorials/state/resource-lifecycle#prevent-resource-deletion
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

	d, err := NewResourceData(a.resource.SchemaInternal, a.resource.IDFields,
		WithPlan(req.Plan),
		WithConfig(req.Config),
	)
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

	ctx, drainWarnings := withWarnings(ctx, diags)
	defer drainWarnings()

	err = a.resource.Create(ctx, a.client, d)
	if err != nil && !(a.resource.IgnoreAlreadyExists && avngen.IsAlreadyExists(err)) {
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

	d, err := NewResourceData(a.resource.SchemaInternal, a.resource.IDFields,
		WithState(req.State),
	)
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

	ctx, drainWarnings := withWarnings(ctx, diags)
	defer drainWarnings()

	// When RemoveMissing is enabled, we remove the resource from the state if it's missing.
	// See ResourceOptions.RemoveMissing for more details.
	err = a.resource.Read(ctx, a.client, d)
	if a.resource.RemoveMissing && IsNotFound(err) {
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

// ErrRefreshStateDesired indicates a RefreshStateDesired attribute did not match after Read.
var ErrRefreshStateDesired = errors.New("resource is not in the desired state")

// defaultRefreshStateRetryDelay is the fixed delay between Read attempts when
// ResourceOptions.refreshStateRetryDelay is zero.
const defaultRefreshStateRetryDelay = time.Second * 5

// refreshState reads the resource after Create or Update with retries, then validates the state
// against ResourceOptions.RefreshStateDesired.
//
// A non-zero ResourceOptions.RefreshStateDelay is waited out before the first attempt (honoring
// ctx). The delay is fixed (refreshStateRetryDelay, default defaultRefreshStateRetryDelay) and there
// is no attempt cap: the poll retries the errors isRefreshStateRetryable accepts until it succeeds
// or ctx (the create/update timeout) is cancelled. On timeout ctx.Err() is returned wrapped with the
// last attempt's error, so the user sees why the state never settled rather than a bare deadline.
func (a *resourceAdapter) refreshState(ctx context.Context, rd ResourceData) error {
	retryDelay := a.resource.refreshStateRetryDelay
	if retryDelay == 0 {
		retryDelay = defaultRefreshStateRetryDelay
	}

	if a.resource.RefreshStateDelay != 0 {
		delay := time.NewTimer(a.resource.RefreshStateDelay)
		defer delay.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-delay.C:
		}
	}

	// The outer warning collector is installed by Create/Update. Each attempt uses a child collector
	// so its warnings are buffered separately; only the last attempt's are merged into the outer one
	// below.
	var lastAttemptWarnings diag.Diagnostics
	err := retry.Do(
		func() error {
			var attemptWarnings diag.Diagnostics
			ctx, drainWarnings := withWarnings(ctx, &attemptWarnings)
			err := a.resource.Read(ctx, a.client, rd)
			drainWarnings()
			lastAttemptWarnings = attemptWarnings

			if err != nil {
				return err
			}

			for key, want := range a.resource.RefreshStateDesired {
				state := rd.Get(key)
				if !Equal(state, want) {
					return fmt.Errorf("%w: expected %q to be `%v`, got `%v`", ErrRefreshStateDesired, key, want, state)
				}
			}
			return nil
		},
		retry.Delay(retryDelay),
		retry.DelayType(retry.FixedDelay),
		// No attempt cap: Attempts(0) retries until success or ctx (the create/update timeout) is
		// cancelled.
		retry.Attempts(0),
		retry.LastErrorOnly(true),
		retry.Context(ctx),
		// On timeout, wrap ctx.Err() with the last attempt's error so the user sees why the state
		// never settled (e.g. the unmet RefreshStateDesired).
		retry.WrapContextErrorWithLastError(true),
		retry.RetryIf(isRefreshStateRetryable),
	)

	AddWarnings(ctx, lastAttemptWarnings...)

	return err
}

// isRefreshStateRetryable reports whether refreshState should retry after err.
//
// Retry on:
//   - 404 Not Found (including ErrNotFound): API may not immediately reflect new/updated resources;
//     resource might not be present in list or direct GET call.
//   - 403 Forbidden: Temporary permission issues can occur due to eventual consistency
//     (e.g., "organization_project").
//   - ErrRefreshStateDesired: A RefreshStateDesired attribute did not match after Read.
func isRefreshStateRetryable(err error) bool {
	return IsNotFound(err) || isAivenError(err, http.StatusForbidden) || errors.Is(err, ErrRefreshStateDesired)
}

func (a *resourceAdapter) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	rsp *resource.UpdateResponse,
) {
	diags := &rsp.Diagnostics

	d, err := NewResourceData(a.resource.SchemaInternal, a.resource.IDFields,
		WithPlan(req.Plan),
		WithState(req.State),
		WithConfig(req.Config),
	)
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

	ctx, drainWarnings := withWarnings(ctx, diags)
	defer drainWarnings()

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
	if a.resource.Delete == nil {
		// Some resources might not have a Delete operation
		return
	}

	diags := &rsp.Diagnostics

	d, err := NewResourceData(a.resource.SchemaInternal, a.resource.IDFields,
		WithState(req.State),
	)
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

	ctx, drainWarnings := withWarnings(ctx, diags)
	defer drainWarnings()

	// The Aiven client might receive 5xx errors from the backend, causing it to retry the delete operation.
	// However, the resource may have already been deleted, in which case a 404 error can be safely ignored.
	//
	// When the delete poller is enabled (DeleteState is set), deleteState owns the Delete call and polls
	// Read until the resource is gone or reaches its terminal state.
	if a.resource.DeleteState != nil {
		err = a.deleteState(ctx, d)
	} else {
		err = a.resource.Delete(ctx, a.client, d)
	}
	if err != nil && !IsNotFound(err) {
		diags.AddError("failed to delete resource", err.Error())
	}
}

// DeleteStateOptions configures the delete poller. When set (non-nil) on ResourceOptions, Delete
// re-issues Delete on every attempt (ignoring its error) and polls Read until the resource reaches
// its terminal state. See deleteState.
type DeleteStateOptions struct {
	// Desired maps attribute names to the terminal value the poller must observe after Delete. Empty
	// means there are no attribute conditions, so the poll completes only once the resource is gone (404).
	Desired map[string]string

	// retryDelay is the fixed delay between attempts. Zero falls back to defaultDeleteStateRetryDelay.
	// Unexported: intended for tests inside the adapter package only.
	retryDelay time.Duration
}

// ErrDeleteStateDesired indicates the delete poller's terminal state was not yet reached after a Read.
var ErrDeleteStateDesired = errors.New("resource has not reached the desired deleted state")

// defaultDeleteStateRetryDelay is the fixed delay between attempts when DeleteStateOptions.retryDelay
// is zero.
const defaultDeleteStateRetryDelay = time.Second * 5

// deleteState deletes the resource and polls until it reaches its terminal state: the API reports it
// gone (404), or every DeleteStateOptions.Desired entry matches (with no Desired, a 404 is the only
// terminal).
//
// Delete is issued until it succeeds once, then each attempt only reads. While Delete errors
// (404 = already gone, 409 = dependents still detaching) it is re-issued, but its error is not fatal
// on its own, so the Read decides the outcome:
//   - a 404 ends the poll and Delete's outer guard treats it as a successful deletion;
//   - any other Read error fails the delete;
//   - a successful Read means the resource is still present, so the poll continues.
//
// The delay is fixed and there is no attempt cap: the poll runs until it succeeds or ctx (the delete
// timeout) is cancelled. On timeout ctx.Err() is returned wrapped with the last attempt's error,
// including the most recent Delete error, so a stuck delete reports why.
//
// Each Read's warnings are buffered so only the last attempt's are merged into the outer collector.
func (a *resourceAdapter) deleteState(ctx context.Context, rd ResourceData) error {
	opts := a.resource.DeleteState
	retryDelay := opts.retryDelay
	if retryDelay == 0 {
		retryDelay = defaultDeleteStateRetryDelay
	}

	var lastAttemptWarnings diag.Diagnostics
	var deleteSucceeded bool
	var deleteErr error
	err := retry.Do(
		func() error {
			// Delete is issued until it succeeds once, then only Read polls. While it errors (e.g.
			// 409 = dependents still detaching) it is re-issued next attempt and the error is
			// remembered; it isn't fatal on its own, so the Read below drives the outcome. The
			// remembered error is attached to the not-terminal error so a timeout surfaces why, and
			// is cleared once Delete finally succeeds.
			if !deleteSucceeded {
				deleteErr = a.resource.Delete(ctx, a.client, rd)
				deleteSucceeded = deleteErr == nil
				if !deleteSucceeded {
					tflog.Debug(ctx, fmt.Sprintf("Delete of %q returned an error (will poll): %s", a.resource.TypeName, deleteErr))
				}
			}

			var attemptWarnings diag.Diagnostics
			ctx, drainWarnings := withWarnings(ctx, &attemptWarnings)
			err := a.resource.Read(ctx, a.client, rd)
			drainWarnings()
			lastAttemptWarnings = attemptWarnings

			// A 404 (gone) bubbles up; Delete's outer guard treats it as success.
			if err != nil {
				return err
			}

			// Still present: with no Desired the only terminal is a 404, so keep polling; otherwise
			// every entry must match. The last Delete error is attached after the loop.
			if len(opts.Desired) == 0 {
				return fmt.Errorf("%w: waiting for the resource to be gone", ErrDeleteStateDesired)
			}
			for key, want := range opts.Desired {
				if state := rd.Get(key); !Equal(state, want) {
					return fmt.Errorf("%w: expected %q to be `%v`, got `%v`", ErrDeleteStateDesired, key, want, state)
				}
			}
			return nil
		},
		retry.Delay(retryDelay),
		retry.DelayType(retry.FixedDelay),
		// No attempt cap: Attempts(0) retries until success or ctx (the delete timeout) is cancelled.
		retry.Attempts(0),
		retry.LastErrorOnly(true),
		retry.Context(ctx),
		// On timeout, wrap ctx.Err() with the last attempt's error so a stuck delete reports why.
		retry.WrapContextErrorWithLastError(true),
		retry.RetryIf(func(err error) bool {
			// Keep polling only while not-terminal; a 404 or any other Read error ends the poll.
			return errors.Is(err, ErrDeleteStateDesired)
		}),
	)

	AddWarnings(ctx, lastAttemptWarnings...)

	// With no attempt cap, the loop only stops on success (nil), a terminal Read error (404 or any
	// other), or context cancellation. So an ErrDeleteStateDesired here can only be the last attempt's
	// error wrapped into the context deadline (retry.WrapContextErrorWithLastError): the poll timed
	// out. Join the last unresolved Delete error (nil once Delete succeeded) so the deadline also
	// reports why the delete never finished. All other outcomes pass through untouched.
	if errors.Is(err, ErrDeleteStateDesired) {
		err = errors.Join(err, deleteErr)
	}

	return err
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

	if a.client == nil {
		// By some reason ValidateConfig is called several times.
		// The first call goes before resourceAdapter.Configure, hence the client is nil.
		return
	}

	diags := &rsp.Diagnostics

	d, err := NewResourceData(a.resource.SchemaInternal, a.resource.IDFields,
		WithConfig(req.Config),
	)
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

	ctx, drainWarnings := withWarnings(ctx, diags)
	defer drainWarnings()

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
	opts := []ResourceDataOpt{
		WithPreservePlanValues(),
		WithPlan(req.Plan),
		WithState(req.State),
		WithConfig(req.Config),
	}

	d, err := NewResourceData(a.resource.SchemaInternal, a.resource.IDFields, opts...)
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

	ctx, drainWarnings := withWarnings(ctx, diags)
	defer drainWarnings()

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

// Equal compares two values for equality, dereferencing pointers as needed.
// Can compare an enum value with its string representation.
func Equal(a, b any) bool {
	return fmt.Sprint(deref(a)) == fmt.Sprint(deref(b))
}

func deref(a any) any {
	if a == nil {
		return nil
	}
	v := reflect.ValueOf(a)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		return v.Elem().Interface()
	}
	return a
}

var (
	ErrNotFound = errors.New("not found")
	ErrMultiple = errors.New("multiple found")
)

// FindOne returns the single element of list for which pred(index) is true. Returns
// ErrNotFound when no element matches; on multiple matches it returns a zero T and an
// error wrapping ErrMultiple with the total count.
func FindOne[T any](list []T, pred func(int) bool) (T, error) {
	var (
		zero  T
		index int
		count int
	)
	for i := range list {
		if !pred(i) {
			continue
		}
		if count == 0 {
			index = i
		}
		count++
	}
	switch count {
	case 0:
		return zero, ErrNotFound
	case 1:
		return list[index], nil
	default:
		return zero, fmt.Errorf("%w: %d", ErrMultiple, count)
	}
}

// IsNotFound reports whether err is a "not found" error: either the local
// ErrNotFound sentinel or an upstream avngen 404.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound) || avngen.IsNotFound(err)
}

func isAivenError(err error, status int) bool {
	var e avngen.Error
	return errors.As(err, &e) && e.Status == status
}
