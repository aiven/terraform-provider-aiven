package organization

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

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
	client *aiven.Client

	// typeName is the name of the resource type.
	typeName string
}

// organizationResourceModel is the model for the organization resource.
type organizationResourceModel struct {
	// ID is the identifier of the organization.
	ID types.String `tfsdk:"id"`
	// Name is the name of the organization.
	Name types.String `tfsdk:"name"`
	// TenantID is the tenant identifier of the organization.
	TenantID types.String `tfsdk:"tenant_id"`
	// CreateTime is the timestamp of the creation of the organization.
	CreateTime types.String `tfsdk:"create_time"`
	// UpdateTime is the timestamp of the last update of the organization.
	UpdateTime types.String `tfsdk:"update_time"`
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
		Description: "Creates and manages an organization in Aiven.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier of the organization.",
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
				Description: "Tenant identifier of the organization.",
				Computed:    true,
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

	client, ok := req.ProviderData.(*aiven.Client)
	if !ok {
		resp.Diagnostics = util.DiagErrorUnexpectedProviderDataType(resp.Diagnostics, req.ProviderData)

		return
	}

	r.client = client
}

// fillModel fills the organization resource model from the Aiven API.
func (r *organizationResource) fillModel(model *organizationResourceModel) (err error) {
	normalizedID, err := schemautil.NormalizeOrganizationID(r.client, model.ID.ValueString())
	if err != nil {
		return
	}

	account, err := r.client.Accounts.Get(normalizedID)
	if err != nil {
		return
	}

	model.Name = types.StringValue(account.Account.Name)

	model.TenantID = types.StringValue(account.Account.TenantId)

	model.CreateTime = types.StringValue(account.Account.CreateTime.String())

	model.UpdateTime = types.StringValue(account.Account.UpdateTime.String())

	return
}

// Create creates an organization resource.
func (r *organizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationResourceModel

	if !util.PlanStateToModel(ctx, &req.Plan, &plan, &resp.Diagnostics) {
		return
	}

	account, err := r.client.Accounts.Create(aiven.Account{
		Name: plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics = util.DiagErrorCreatingResource(resp.Diagnostics, r, err)

		return
	}

	plan.ID = types.StringValue(account.Account.OrganizationId)

	err = r.fillModel(&plan)
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

	err := r.fillModel(&state)
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

	normalizedID, err := schemautil.NormalizeOrganizationID(r.client, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics = util.DiagErrorUpdatingResource(resp.Diagnostics, r, err)

		return
	}

	_, err = r.client.Accounts.Update(normalizedID, aiven.Account{
		Name: plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics = util.DiagErrorUpdatingResource(resp.Diagnostics, r, err)

		return
	}

	err = r.fillModel(&plan)
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

	normalizedID, err := schemautil.NormalizeOrganizationID(r.client, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics = util.DiagErrorDeletingResource(resp.Diagnostics, r, err)

		return
	}

	err = r.client.Accounts.Delete(normalizedID)
	if err != nil {
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
