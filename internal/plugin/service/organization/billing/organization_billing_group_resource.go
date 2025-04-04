package billing

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationbilling"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/diagnostics"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	providertypes "github.com/aiven/terraform-provider-aiven/internal/plugin/types"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/validation"
)

var (
	_ resource.Resource                = &organizationBillingGroupResource{}
	_ resource.ResourceWithConfigure   = &organizationBillingGroupResource{}
	_ resource.ResourceWithImportState = &organizationBillingGroupResource{}
)

// NewOrganizationBillingGroupResource is a constructor for the organization billing group resource.
func NewOrganizationBillingGroupResource() resource.Resource {
	return &organizationBillingGroupResource{}
}

// organizationBillingGroupResource is the organization billing group resource implementation.
type organizationBillingGroupResource struct {
	// client is the instance of the Aiven client to use.
	client avngen.Client

	// diag is the diagnostics helper.
	diag *diagnostics.DiagnosticsHelper
}

// TypeName returns the resource type name.
func (r *organizationBillingGroupResource) TypeName() string {
	return "organization_billing_group_resource"
}

// Metadata returns the metadata for the organization billing group resource.
func (r *organizationBillingGroupResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_organization_billing_group"
}

// Schema defines the schema for the organization billing group resource.
func (r *organizationBillingGroupResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = billingResourceSchema(ctx)
}

// Configure sets up the organization billing group resource.
func (r *organizationBillingGroupResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	p, ok := req.ProviderData.(providertypes.AivenClientProvider)
	if !ok {
		resp.Diagnostics.AddError(
			errmsg.SummaryUnexpectedProviderDataType,
			fmt.Sprintf(errmsg.DetailUnexpectedProviderDataType, req.ProviderData),
		)
		return
	}

	r.client = p.GetGenClient()
	r.diag = diagnostics.NewDiagnosticsHelper(r.TypeName())
}

// getEmailStrings extracts email strings from BillingContactEmailOut or BillingEmailOut slices
func getEmailStrings(emails interface{}) []string {
	switch e := emails.(type) {
	case []organizationbilling.BillingContactEmailOut:
		result := make([]string, len(e))
		for i, email := range e {
			result[i] = email.Email
		}
		return result
	case []organizationbilling.BillingEmailOut:
		result := make([]string, len(e))
		for i, email := range e {
			result[i] = email.Email
		}
		return result
	default:
		return nil
	}
}

// Create creates an organization billing group resource.
func (r *organizationBillingGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModelBilling

	// Get plan values
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var createReq organizationbilling.OrganizationBillingGroupCreateIn
	diags := expandModelBilling(ctx, &plan.baseModelBilling, &createReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the billing group
	billingGroup, err := r.client.OrganizationBillingGroupCreate(ctx, plan.OrganizationID.ValueString(), &createReq)
	if err != nil {
		r.diag.AddError(&resp.Diagnostics, "creating", err)
		return
	}

	// Set state
	resp.Diagnostics.Append(flattenModelBilling(ctx, &plan.baseModelBilling, billingGroup)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads an organization billing group resource.
func (r *organizationBillingGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModelBilling

	// Get state values
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the billing group
	billingGroup, err := r.client.OrganizationBillingGroupGet(ctx, state.OrganizationID.ValueString(), state.BillingGroupID.ValueString())
	if err != nil {
		r.diag.AddError(&resp.Diagnostics, "reading", err)
		return
	}

	// Set state
	resp.Diagnostics.Append(flattenModelBilling(ctx, &state.baseModelBilling, billingGroup)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates an organization billing group resource.
func (r *organizationBillingGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and state
	var plan, state resourceModelBilling
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var updateReq organizationbilling.OrganizationBillingGroupUpdateIn
	diags := expandModelBilling(ctx, &plan.baseModelBilling, &updateReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the billing group
	billingGroup, err := r.client.OrganizationBillingGroupUpdate(
		ctx, plan.OrganizationID.ValueString(), state.BillingGroupID.ValueString(), &updateReq)
	if err != nil {
		r.diag.AddError(&resp.Diagnostics, "updating", err)
		return
	}

	// Set state
	resp.Diagnostics.Append(flattenModelBilling(ctx, &plan.baseModelBilling, billingGroup)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes an organization billing group resource.
func (r *organizationBillingGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceModelBilling

	// Get state values
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the billing group
	err := r.client.OrganizationBillingGroupDelete(ctx, state.OrganizationID.ValueString(), state.BillingGroupID.ValueString())
	if err != nil {
		r.diag.AddError(&resp.Diagnostics, "deleting", err)
		return
	}
}

// ImportState imports an organization billing group resource.
func (r *organizationBillingGroupResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Check if import ID is valid
	parts, err := validation.ValidateImportID(req.ID, "organization_id/billing_group_id")
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}

	// Set the import attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("billing_group_id"), parts[1])...)
}
