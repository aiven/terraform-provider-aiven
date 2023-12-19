package organization

import (
	"context"
	"fmt"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &organizationUserGroupMembersResource{}
	_ resource.ResourceWithConfigure   = &organizationUserGroupMembersResource{}
	_ resource.ResourceWithImportState = &organizationUserGroupMembersResource{}
)

// NewOrganizationUserGroupMembersResource is a constructor for the organization resource.
func NewOrganizationUserGroupMembersResource() resource.Resource {
	return &organizationUserGroupMembersResource{}
}

// organizationUserGroupMembersResource is the organization resource implementation.
type organizationUserGroupMembersResource struct {
	// client is the instance of the Aiven client to use.
	client *aiven.Client

	// typeName is the name of the resource type.
	typeName string
}

// organizationUserGroupMembersResourceModel is the model for the organization resource.
type organizationUserGroupMembersResourceModel struct {
	// ID is the identifier of the organization.
	OrganizationID types.String `tfsdk:"organization_id"`

	// ID is the identifier of the organization user group.
	OrganizationGroupID types.String `tfsdk:"group_id"`

	// ID is the identifier of the organization user group member.
	OrganizationUserID types.String `tfsdk:"user_id"`

	// Last activity time of the user group member.
	LastActivityTime types.String `tfsdk:"last_activity_time"`

	// Timeouts is the configuration for resource-specific timeouts.
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

// Metadata returns the metadata for the organization resource.
func (r *organizationUserGroupMembersResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_organization_user_group_member"

	r.typeName = resp.TypeName
}

func (r *organizationUserGroupMembersResource) TypeName() string {
	return r.typeName
}

// Schema returns the schema for the organization resource.
func (r *organizationUserGroupMembersResource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = util.GeneralizeSchema(ctx, schema.Schema{
		Description: "Creates and manages an organization user group members in Aiven. " +
			"Please no that this resource is " + " in beta and may change without notice. " +
			"To use it please use the beta environment variable PROVIDER_AIVEN_ENABLE_BETA.",
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.StringAttribute{
				Description: "Identifier of the organization.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				Description: "Identifier of the organization user group member.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group_id": schema.StringAttribute{
				Description: "Identifier of the organization user group.",
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

// Configure configures the organization resource.
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

// TimeoutSchema returns the schema for resource-specific timeouts.
func (r *organizationUserGroupMembersResource) fillModel(
	ctx context.Context,
	model *organizationUserGroupMembersResourceModel,
) error {
	list, err := r.client.OrganizationUserGroupMembers.List(
		ctx,
		model.OrganizationID.ValueString(),
		model.OrganizationGroupID.ValueString())
	if err != nil {
		return err
	}

	if len(list.Members) == 0 {
		return nil
	}

	member := list.Members[0]
	model.LastActivityTime = types.StringValue(member.LastActivityTime.String())

	return nil
}

// Create creates an organization resource.
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
		plan.OrganizationGroupID.ValueString(),
		aiven.OrganizationUserGroupMemberRequest{
			Operation: "add_members",
			MemberIDs: []string{
				plan.OrganizationUserID.ValueString(),
			}})
	if err != nil {
		resp.Diagnostics = util.DiagErrorCreatingResource(resp.Diagnostics, r, err)

		return
	}

	err = r.fillModel(ctx, &plan)
	if err != nil {
		resp.Diagnostics = util.DiagErrorCreatingResource(resp.Diagnostics, r, err)

		return
	}

	if !util.ModelToPlanState(ctx, plan, &resp.State, &resp.Diagnostics) {
		return
	}
}

// Delete deletes an organization resource.
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

// Delete deletes an organization resource.
func (r *organizationUserGroupMembersResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var plan organizationUserGroupMembersResourceModel

	if !util.PlanStateToModel(ctx, &req.State, &plan, &resp.Diagnostics) {
		return
	}

	err := r.client.OrganizationUserGroupMembers.Modify(
		ctx,
		plan.OrganizationID.ValueString(),
		plan.OrganizationGroupID.ValueString(),
		aiven.OrganizationUserGroupMemberRequest{
			Operation: "remove_members",
			MemberIDs: []string{
				plan.OrganizationGroupID.ValueString(),
			}})
	if err != nil {
		resp.Diagnostics = util.DiagErrorDeletingResource(resp.Diagnostics, r, err)

		return
	}
}

// Update updates an organization resource.
func (r *organizationUserGroupMembersResource) Update(
	_ context.Context,
	_ resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	util.DiagErrorUpdatingResource(
		resp.Diagnostics,
		r,
		fmt.Errorf("cannot update %s resource", r.TypeName()),
	)
}

// ImportStatePassthroughID is a helper function to set the import
func (r *organizationUserGroupMembersResource) ImportState(
	_ context.Context,
	_ resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	util.DiagErrorUpdatingResource(
		resp.Diagnostics,
		r,
		fmt.Errorf("cannot import %s resource", r.TypeName()),
	)
}
