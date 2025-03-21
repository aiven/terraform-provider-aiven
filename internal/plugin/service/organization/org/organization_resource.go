package org

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/account"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/diagnostics"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	providertypes "github.com/aiven/terraform-provider-aiven/internal/plugin/types"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

var (
	_ resource.Resource                = &organizationResource{}
	_ resource.ResourceWithConfigure   = &organizationResource{}
	_ resource.ResourceWithImportState = &organizationResource{}
	_ util.TypeNameable                = &organizationResource{}
)

// NewOrganizationResource is a constructor for the organization resource.
func NewOrganizationResource() resource.Resource {
	return &organizationResource{}
}

// organizationResource is the organization resource implementation.
type organizationResource struct {
	// client is the instance of the Aiven client to use.
	client avngen.Client

	// diag is the diagnostics helper.
	diag *diagnostics.DiagnosticsHelper
}

// organizationResourceModel is the model for the organization resource.
type organizationResourceModel struct {
	OrganizationModel

	// Timeouts is the configuration for resource-specific timeouts.
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

// TypeName returns the resource type name for the organization resource.
func (r *organizationResource) TypeName() string {
	return "organization_resource"
}

// Metadata returns the metadata for the organization resource.
func (r *organizationResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

// Schema defines the schema for the organization resource.
func (r *organizationResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	schemaObj := schema.Schema{
		Description: "Creates and manages an [organization](https://aiven.io/docs/platform/concepts/orgs-units-projects).",
		Attributes:  ResourceSchema(),
	}

	// Add timeouts block
	if schemaObj.Blocks == nil {
		schemaObj.Blocks = make(map[string]schema.Block)
	}
	schemaObj.Blocks["timeouts"] = timeouts.BlockAll(ctx)

	resp.Schema = schemaObj
}

// Configure sets up the organization resource.
func (r *organizationResource) Configure(
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

// Create creates an organization resource.
func (r *organizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create organization
	created, err := r.client.AccountCreate(ctx, &account.AccountCreateIn{
		AccountName: plan.Name.ValueString(),
	})
	if err != nil {
		r.diag.AddError(&resp.Diagnostics, "creating", err)
		return
	}

	// Get the organization details
	account, err := r.client.AccountGet(ctx, created.AccountId)
	if err != nil {
		r.diag.AddError(&resp.Diagnostics, "reading", err)
		return
	}

	// Set all fields - store organization ID in state
	plan.ID = types.StringValue(account.OrganizationId)
	plan.Name = types.StringValue(account.AccountName)
	plan.TenantID = types.StringPointerValue(account.TenantId)
	plan.CreateTime = types.StringValue(account.CreateTime.String())
	plan.UpdateTime = types.StringValue(account.UpdateTime.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read reads an organization resource.
func (r *organizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve the organization ID to account ID for API call
	accountID, err := ResolveAccountID(ctx, r.client, state.ID.ValueString())
	if err != nil {
		if avngen.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		r.diag.AddError(&resp.Diagnostics, "resolving account ID", err)
		return
	}

	// Get the organization using the account ID
	account, err := r.client.AccountGet(ctx, accountID)
	if err != nil {
		if avngen.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		r.diag.AddError(&resp.Diagnostics, "reading", err)
		return
	}

	// Set all fields - keep organization ID in state
	state.ID = types.StringValue(account.OrganizationId)
	state.Name = types.StringValue(account.AccountName)
	state.TenantID = types.StringPointerValue(account.TenantId)
	state.CreateTime = types.StringValue(account.CreateTime.String())
	state.UpdateTime = types.StringValue(account.UpdateTime.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update updates an organization resource.
func (r *organizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state organizationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve the organization ID to account ID for API call
	accountID, err := ResolveAccountID(ctx, r.client, state.ID.ValueString())
	if err != nil {
		r.diag.AddError(&resp.Diagnostics, "resolving account ID", err)
		return
	}

	// Update name if changed
	name := plan.Name.ValueString()
	if _, err := r.client.AccountUpdate(ctx, accountID, &account.AccountUpdateIn{
		AccountName: &name,
	}); err != nil {
		r.diag.AddError(&resp.Diagnostics, "updating", err)
		return
	}

	// Read the resource after update
	account, err := r.client.AccountGet(ctx, accountID)
	if err != nil {
		r.diag.AddError(&resp.Diagnostics, "reading", err)
		return
	}

	// Set all fields - keep organization ID in state
	plan.ID = types.StringValue(account.OrganizationId)
	plan.Name = types.StringValue(account.AccountName)
	plan.TenantID = types.StringPointerValue(account.TenantId)
	plan.CreateTime = types.StringValue(account.CreateTime.String())
	plan.UpdateTime = types.StringValue(account.UpdateTime.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes an organization resource.
func (r *organizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve the organization ID to account ID for API call
	accountID, err := ResolveAccountID(ctx, r.client, state.ID.ValueString())
	if err != nil {
		if avngen.IsNotFound(err) {
			// Resource already gone, nothing to do
			return
		}
		r.diag.AddError(&resp.Diagnostics, "resolving account ID", err)
		return
	}

	// Delete the organization using the account ID
	if err := r.client.AccountDelete(ctx, accountID); err != nil {
		if !avngen.IsNotFound(err) {
			r.diag.AddError(&resp.Diagnostics, "deleting", err)
			return
		}
	}
}

// ImportState handles resource's state import requests.
func (r *organizationResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Import the ID directly - it will be resolved to account ID during Read operations
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
