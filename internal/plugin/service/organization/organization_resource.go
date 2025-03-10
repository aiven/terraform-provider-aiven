package organization

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	account2 "github.com/aiven/go-client-codegen/handler/account"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

var (
	_ resource.Resource                = &organizationResource{}
	_ resource.ResourceWithConfigure   = &organizationResource{}
	_ resource.ResourceWithImportState = &organizationResource{}

	_ util.TypeNameable = &organizationResource{}
)

// NewOrganizationResource is a constructor for the organization resource.
func NewOrganizationResource() resource.Resource {
	return &organizationResource{}
}

// organizationResource is the organization resource implementation.
type organizationResource struct {
	// client is the instance of the Aiven client to use.
	client avngen.Client

	// typeName is the name of the resource type.
	typeName string
}

// organizationResourceModel is the model for the organization resource.
type organizationResourceModel struct {
	organizationDataSourceModel
	// Timeouts is the configuration for resource-specific timeouts.
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

// Metadata returns the metadata for the organization resource.
func (r *organizationResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_organization"

	r.typeName = resp.TypeName
}

// TypeName returns the resource type name for the organization resource.
func (r *organizationResource) TypeName() string {
	return r.typeName
}

// Schema defines the schema for the organization resource.
func (r *organizationResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = util.GeneralizeSchema(ctx, schema.Schema{
		Description: "Creates and manages an [organization](https://aiven.io/docs/platform/concepts/orgs-units-projects).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the organization.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the organization.",
				Required:    true,
			},
			"tenant_id": schema.StringAttribute{
				Description: "Tenant ID of the organization.",
				Computed:    true,
			},
			"parent_account_id": schema.StringAttribute{
				Description: "ID of the parent account of the organization.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"primary_billing_group_id": schema.StringAttribute{
				Description: "ID of the primary billing group of the organization.",
				Computed:    true,
				Optional:    true,
			},
			"create_time": schema.StringAttribute{
				Description: "Timestamp of the creation of the organization.",
				Computed:    true,
			},
			"update_time": schema.StringAttribute{
				Description: "Timestamp of the last update of the organization.",
				Computed:    true,
			},
		},
	})
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

	client, err := common.GenClient()
	if err != nil {
		resp.Diagnostics.AddError(errmsg.SummaryUnexpectedProviderDataType, err.Error())
		return
	}

	r.client = client
}

// fillModel fills the organization resource model from the Aiven API.
func (r *organizationResource) fillModel(ctx context.Context, model *organizationResourceModel) (err error) {
	normalizedID, err := schemautil.NormalizeOrganizationIDGen(ctx, r.client, model.ID.ValueString())
	if err != nil {
		return
	}

	account, err := r.client.AccountGet(ctx, normalizedID)
	if err != nil {
		return
	}

	model.Name = types.StringValue(account.AccountName)
	model.TenantID = types.StringPointerValue(account.TenantId)
	model.ParentAccountID = types.StringPointerValue(account.ParentAccountId)
	model.PrimaryBillingGroupID = types.StringValue(account.PrimaryBillingGroupId)
	model.CreateTime = types.StringValue(account.CreateTime.String())
	model.UpdateTime = types.StringValue(account.UpdateTime.String())
	return
}

// Create creates an organization resource.
func (r *organizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationResourceModel

	if !util.PlanStateToModel(ctx, &req.Plan, &plan, &resp.Diagnostics) {
		return
	}

	in := &account2.AccountCreateIn{
		AccountName: plan.Name.ValueString(),
	}

	if !plan.PrimaryBillingGroupID.IsUnknown() {
		in.PrimaryBillingGroupId = plan.PrimaryBillingGroupID.ValueStringPointer()
	}

	if !plan.ParentAccountID.IsUnknown() {
		in.PrimaryBillingGroupId = plan.ParentAccountID.ValueStringPointer()
	}

	account, err := r.client.AccountCreate(ctx, in)
	if err != nil {
		resp.Diagnostics = util.DiagErrorCreatingResource(resp.Diagnostics, r, err)

		return
	}

	plan.ID = types.StringValue(account.OrganizationId)

	err = r.fillModel(ctx, &plan)
	if err != nil {
		resp.Diagnostics = util.DiagErrorCreatingResource(resp.Diagnostics, r, err)

		return
	}

	if !util.ModelToPlanState(ctx, plan, &resp.State, &resp.Diagnostics) {
		return
	}
}

// Read reads an organization resource.
func (r *organizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationResourceModel

	if !util.PlanStateToModel(ctx, &req.State, &state, &resp.Diagnostics) {
		return
	}

	err := r.fillModel(ctx, &state)
	if err != nil {
		resp.Diagnostics = util.DiagErrorReadingResource(resp.Diagnostics, r, err)

		return
	}

	if !util.ModelToPlanState(ctx, state, &resp.State, &resp.Diagnostics) {
		return
	}
}

// Update updates an organization resource.
func (r *organizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan organizationResourceModel

	if !util.PlanStateToModel(ctx, &req.Plan, &plan, &resp.Diagnostics) {
		return
	}

	normalizedID, err := schemautil.NormalizeOrganizationIDGen(ctx, r.client, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics = util.DiagErrorUpdatingResource(resp.Diagnostics, r, err)

		return
	}

	in := &account2.AccountUpdateIn{
		AccountName: plan.Name.ValueStringPointer(),
	}

	if !plan.PrimaryBillingGroupID.IsUnknown() {
		in.PrimaryBillingGroupId = plan.PrimaryBillingGroupID.ValueStringPointer()
	}

	if _, err = r.client.AccountUpdate(ctx, normalizedID, in); err != nil {
		resp.Diagnostics = util.DiagErrorUpdatingResource(resp.Diagnostics, r, err)

		return
	}

	err = r.fillModel(ctx, &plan)
	if err != nil {
		resp.Diagnostics = util.DiagErrorUpdatingResource(resp.Diagnostics, r, err)

		return
	}

	if !util.ModelToPlanState(ctx, plan, &resp.State, &resp.Diagnostics) {
		return
	}
}

// Delete deletes an organization resource.
func (r *organizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationResourceModel

	if !util.PlanStateToModel(ctx, &req.State, &state, &resp.Diagnostics) {
		return
	}

	normalizedID, err := schemautil.NormalizeOrganizationIDGen(ctx, r.client, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics = util.DiagErrorDeletingResource(resp.Diagnostics, r, err)

		return
	}

	if err = r.client.AccountDelete(ctx, normalizedID); err != nil {
		resp.Diagnostics = util.DiagErrorDeletingResource(resp.Diagnostics, r, err)

		return
	}
}

// ImportState handles resource's state import requests.
func (r *organizationResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
