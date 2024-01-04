package organization

import (
	"context"
	"fmt"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

var (
	_ resource.Resource                = &organizationApplicationUser{}
	_ resource.ResourceWithConfigure   = &organizationApplicationUser{}
	_ resource.ResourceWithImportState = &organizationApplicationUser{}

	_ util.TypeNameable = &organizationApplicationUser{}
)

// NewOrganizationApplicationUser is a constructor for the organization application user resource.
func NewOrganizationApplicationUser() resource.Resource {
	return &organizationApplicationUser{}
}

// organizationApplicationUser is the organization application user resource implementation.
type organizationApplicationUser struct {
	// client is the instance of the Aiven client to use.
	client *aiven.Client

	// typeName is the name of the resource type.
	typeName string
}

// organizationApplicationUserModel is the model for the organization application user resource.
type organizationApplicationUserModel struct {
	// ID is the identifier of the organization application user.
	ID types.String `tfsdk:"id"`
	// UserID is the identifier of the organization application user.
	UserID types.String `tfsdk:"user_id"`
	// OrganizationID is the identifier of the organization the application user belongs to.
	OrganizationID types.String `tfsdk:"organization_id"`
	// Name is the name of the organization application user.
	Name types.String `tfsdk:"name"`
	// Email is the email of the organization application user.
	Email types.String `tfsdk:"email"`
	// Timeouts is the configuration for resource-specific timeouts.
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

// Metadata returns the metadata for the organization application user resource.
func (r *organizationApplicationUser) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_organization_application_user"

	r.typeName = resp.TypeName
}

// TypeName returns the resource type name for the organization application user resource.
func (r *organizationApplicationUser) TypeName() string {
	return r.typeName
}

// Schema defines the schema for the organization application user resource.
func (r *organizationApplicationUser) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = util.GeneralizeSchema(ctx, schema.Schema{
		Description: util.BetaDescription("Creates and manages an organization application user in Aiven."),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Compound identifier of the organization application user.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Description: "Identifier of the organization application user.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "Identifier of the organization the application user belongs to.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the organization application user.",
				Required:    true,
			},
			"email": schema.StringAttribute{
				Description: "Email of the organization application user.",
				Computed:    true,
			},
		},
	})
}

// Configure sets up the organization application user resource.
func (r *organizationApplicationUser) Configure(
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

// fillModel fills the organization application user resource model from the Aiven API.
func (r *organizationApplicationUser) fillModel(
	ctx context.Context,
	model *organizationApplicationUserModel,
) (err error) {
	appUsers, err := r.client.OrganizationApplicationUserHandler.List(ctx, model.OrganizationID.ValueString())
	if err != nil {
		return err
	}

	var appUser *aiven.ApplicationUserInfo

	for _, u := range appUsers.Users {
		if u.UserID == model.UserID.ValueString() {
			appUser = &u
			break
		}
	}

	if appUser == nil {
		return fmt.Errorf(errmsg.AivenResourceNotFound, r.TypeName(), model.ID.ValueString())
	}

	model.Name = types.StringValue(appUser.Name)

	model.Email = types.StringValue(appUser.UserEmail)

	return err
}

// Create creates an organization application user resource.
func (r *organizationApplicationUser) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan organizationApplicationUserModel

	if !util.PlanStateToModel(ctx, &req.Plan, &plan, &resp.Diagnostics) {
		return
	}

	appUser, err := r.client.OrganizationApplicationUserHandler.Create(
		ctx,
		plan.OrganizationID.ValueString(),
		aiven.ApplicationUserCreateRequest{
			Name: plan.Name.ValueString(),
		},
	)
	if err != nil {
		resp.Diagnostics = util.DiagErrorCreatingResource(resp.Diagnostics, r, err)

		return
	}

	plan.ID = types.StringValue(util.ComposeID(plan.OrganizationID.ValueString(), appUser.UserID))
	plan.UserID = types.StringValue(appUser.UserID)

	err = r.fillModel(ctx, &plan)
	if err != nil {
		resp.Diagnostics = util.DiagErrorCreatingResource(resp.Diagnostics, r, err)

		return
	}

	if !util.ModelToPlanState(ctx, plan, &resp.State, &resp.Diagnostics) {
		return
	}
}

// Read reads an organization application user resource.
func (r *organizationApplicationUser) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state organizationApplicationUserModel

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
func (r *organizationApplicationUser) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan organizationApplicationUserModel

	if !util.PlanStateToModel(ctx, &req.Plan, &plan, &resp.Diagnostics) {
		return
	}

	if _, err := r.client.OrganizationUser.Update(
		ctx,
		plan.OrganizationID.ValueString(),
		plan.UserID.ValueString(),
		aiven.OrganizationUserUpdateRequest{
			RealName: util.Ref(plan.Name.ValueString()),
		},
	); err != nil {
		resp.Diagnostics = util.DiagErrorUpdatingResource(resp.Diagnostics, r, err)

		return
	}

	err := r.fillModel(ctx, &plan)
	if err != nil {
		resp.Diagnostics = util.DiagErrorUpdatingResource(resp.Diagnostics, r, err)

		return
	}

	if !util.ModelToPlanState(ctx, plan, &resp.State, &resp.Diagnostics) {
		return
	}
}

// Delete deletes an organization application user resource.
func (r *organizationApplicationUser) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state organizationApplicationUserModel

	if !util.PlanStateToModel(ctx, &req.State, &state, &resp.Diagnostics) {
		return
	}

	if err := r.client.OrganizationApplicationUserHandler.Delete(
		ctx,
		state.OrganizationID.ValueString(),
		state.UserID.ValueString(),
	); err != nil {
		resp.Diagnostics = util.DiagErrorDeletingResource(resp.Diagnostics, r, err)

		return
	}
}

// ImportState handles resource's state import requests.
func (r *organizationApplicationUser) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	util.UnpackCompoundID(ctx, req, resp, "organization_id", "user_id")
}
