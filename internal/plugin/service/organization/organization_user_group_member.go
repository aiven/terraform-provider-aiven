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
	_ resource.Resource                = &organizationUserGroupMembersResource{}
	_ resource.ResourceWithConfigure   = &organizationUserGroupMembersResource{}
	_ resource.ResourceWithImportState = &organizationUserGroupMembersResource{}
)

// NewOrganizationUserGroupMembersResource is a constructor for the organization user group member resource.
func NewOrganizationUserGroupMembersResource() resource.Resource {
	return &organizationUserGroupMembersResource{}
}

// organizationUserGroupMembersResource is the organization user group member resource implementation.
type organizationUserGroupMembersResource struct {
	// client is the instance of the Aiven client to use.
	client *aiven.Client

	// typeName is the name of the resource type.
	typeName string
}

// organizationUserGroupMembersResourceModel is the model for the organization user group member resource.
type organizationUserGroupMembersResourceModel struct {
	// ID is the compound identifier of the organization user group member.
	ID types.String `tfsdk:"id"`
	// OrganizationID is the identifier of the organization.
	OrganizationID types.String `tfsdk:"organization_id"`
	// GroupID is the identifier of the organization user group.
	GroupID types.String `tfsdk:"group_id"`
	// UserID is the identifier of the organization user group member.
	UserID types.String `tfsdk:"user_id"`
	// Last activity time of the user group member.
	LastActivityTime types.String `tfsdk:"last_activity_time"`
	// Timeouts is the configuration for resource-specific timeouts.
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

// Metadata returns the metadata for the organization user group member resource.
func (r *organizationUserGroupMembersResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_organization_user_group_member"

	r.typeName = resp.TypeName
}

// TypeName returns the resource type name.
func (r *organizationUserGroupMembersResource) TypeName() string {
	return r.typeName
}

// Schema returns the schema for the organization user group member resource.
func (r *organizationUserGroupMembersResource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = util.GeneralizeSchema(ctx, schema.Schema{
		Description: util.BetaDescription("Adds and manages users in a [user group](https://aiven.io/docs/platform/concepts/projects_accounts_access#groups)."),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "A compound identifier of the group member in the format `organization_id/group_id/user_id`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the organization.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group_id": schema.StringAttribute{
				Description: "The ID of the user group.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				Description: "The ID of the organization user.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"last_activity_time": schema.StringAttribute{
				Description: "Last activity time of the user group member.",
				Computed:    true,
			},
		},
	})
}

// Configure configures the organization user group member resource.
func (r *organizationUserGroupMembersResource) Configure(
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

// fillModel fills the organization group project relation model from the Aiven API.
func (r *organizationUserGroupMembersResource) fillModel(
	ctx context.Context,
	model *organizationUserGroupMembersResourceModel,
) error {
	list, err := r.client.OrganizationUserGroupMembers.List(
		ctx,
		model.OrganizationID.ValueString(),
		model.GroupID.ValueString(),
	)
	if err != nil {
		return err
	}

	if len(list.Members) == 0 {
		return nil
	}

	var member *aiven.OrganizationUserGroupMember

	for _, m := range list.Members {
		if m.UserID == model.UserID.ValueString() {
			member = &m
			break
		}
	}

	if member == nil {
		return fmt.Errorf(
			errmsg.AivenResourceNotFound,
			r.TypeName(),
			util.ComposeID(
				model.OrganizationID.ValueString(),
				model.GroupID.ValueString(),
				model.UserID.ValueString(),
			),
		)
	}

	model.LastActivityTime = util.ValueOrDefault(member.LastActivityTime, types.StringNull())

	return nil
}

// Create creates an organization user group member resource.
func (r *organizationUserGroupMembersResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan organizationUserGroupMembersResourceModel

	if !util.PlanStateToModel(ctx, &req.Plan, &plan, &resp.Diagnostics) {
		return
	}

	err := r.client.OrganizationUserGroupMembers.Modify(
		ctx,
		plan.OrganizationID.ValueString(),
		plan.GroupID.ValueString(),
		aiven.OrganizationUserGroupMemberRequest{
			Operation: "add_members",
			MemberIDs: []string{
				plan.UserID.ValueString(),
			},
		},
	)
	if err != nil {
		resp.Diagnostics = util.DiagErrorCreatingResource(resp.Diagnostics, r, err)

		return
	}

	plan.ID = types.StringValue(
		util.ComposeID(plan.OrganizationID.ValueString(), plan.GroupID.ValueString(), plan.UserID.ValueString()),
	)

	err = r.fillModel(ctx, &plan)
	if err != nil {
		resp.Diagnostics = util.DiagErrorCreatingResource(resp.Diagnostics, r, err)

		return
	}

	if !util.ModelToPlanState(ctx, plan, &resp.State, &resp.Diagnostics) {
		return
	}
}

// Delete deletes an organization user group member resource.
func (r *organizationUserGroupMembersResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state organizationUserGroupMembersResourceModel

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

// Update updates an organization user group member resource.
func (r *organizationUserGroupMembersResource) Update(
	_ context.Context,
	_ resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	resp.Diagnostics = util.DiagErrorUpdatingResourceNotSupported(resp.Diagnostics, r)
}

// Delete deletes an organization user group member resource.
func (r *organizationUserGroupMembersResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var plan organizationUserGroupMembersResourceModel

	if !util.PlanStateToModel(ctx, &req.State, &plan, &resp.Diagnostics) {
		return
	}

	if err := r.client.OrganizationUserGroupMembers.Modify(
		ctx,
		plan.OrganizationID.ValueString(),
		plan.GroupID.ValueString(),
		aiven.OrganizationUserGroupMemberRequest{
			Operation: "remove_members",
			MemberIDs: []string{
				plan.UserID.ValueString(),
			},
		},
	); err != nil {
		resp.Diagnostics = util.DiagErrorDeletingResource(resp.Diagnostics, r, err)

		return
	}
}

// ImportState handles resource's state import requests.
func (r *organizationUserGroupMembersResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	util.UnpackCompoundID(ctx, req, resp, "organization_id", "group_id", "user_id")
}
