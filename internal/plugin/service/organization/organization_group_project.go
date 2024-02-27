package organization

import (
	"context"
	"fmt"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

var (
	_ resource.Resource                = &organizationGroupProjectResource{}
	_ resource.ResourceWithConfigure   = &organizationGroupProjectResource{}
	_ resource.ResourceWithImportState = &organizationGroupProjectResource{}

	_ util.TypeNameable = &organizationGroupProjectResource{}
)

// NewOrganizationGroupProjectResource is a constructor for the organization group project relation resource.
func NewOrganizationGroupProjectResource() resource.Resource {
	return &organizationGroupProjectResource{}
}

// organizationGroupUserResource is the organization group project relation resource implementation.
type organizationGroupProjectResource struct {
	// client is the instance of the Aiven client to use.
	client *aiven.Client

	// typeName is the name of the resource type.
	typeName string
}

// organizationGroupProjectResourceModel is the model for the organization group project relation resource.
type organizationGroupProjectResourceModel struct {
	// ID is the compound identifier of the organization group project relation.
	ID types.String `tfsdk:"id"`
	// Project is the name of the project.
	Project types.String `tfsdk:"project"`
	// GroupID is the identifier of the organization group.
	GroupID types.String `tfsdk:"group_id"`
	// Role is the role of the organization group project relation.
	Role types.String `tfsdk:"role"`
	// Timeouts is the configuration for resource-specific timeouts.
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

// Metadata returns the metadata for the organization group project relation resource.
func (r *organizationGroupProjectResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_organization_group_project"

	r.typeName = resp.TypeName
}

// TypeName returns the resource type name.
func (r *organizationGroupProjectResource) TypeName() string {
	return r.typeName
}

// Schema returns the schema for the resource.
func (r *organizationGroupProjectResource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse) {
	resp.Schema = util.GeneralizeSchema(ctx, schema.Schema{
		Description: util.BetaDescription(
			`Adds and manages a [group](https://aiven.io/docs/platform/concepts/projects_accounts_access#groups) 
			of users as [members of a project](https://aiven.io/docs/platform/reference/project-member-privileges).`,
		),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "A compound identifier of the resource in the format `project/group_id`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project": schema.StringAttribute{
				Description: "The project that the users in the group are members of.",
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
			"role": schema.StringAttribute{
				Description: "Role assigned to all users in the group for the project.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("admin", "developer", "operator", "read_only"),
				},
			},
		},
	})
}

// Configure is called to configure the resource.
func (r *organizationGroupProjectResource) Configure(
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
func (r *organizationGroupProjectResource) fillModel(
	ctx context.Context,
	model *organizationGroupProjectResourceModel,
) error {
	list, err := r.client.ProjectOrganization.List(ctx, model.Project.ValueString())
	if err != nil {
		return err
	}

	var group *aiven.ProjectUserGroup

	for _, g := range list {
		if g.OrganizationGroupID == model.GroupID.ValueString() {
			group = g
			break
		}
	}

	if group == nil {
		return fmt.Errorf(
			errmsg.AivenResourceNotFound,
			r.TypeName(),
			util.ComposeID(model.Project.ValueString(), model.GroupID.ValueString()),
		)
	}

	model.GroupID = types.StringValue(group.OrganizationGroupID)

	model.Role = types.StringValue(group.Role)

	// There is no API endpoint to get the permissions of the organization group project relation.

	return nil
}

// Create creates an organization group project relation resource.
func (r *organizationGroupProjectResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan organizationGroupProjectResourceModel

	if !util.PlanStateToModel(ctx, &req.Plan, &plan, &resp.Diagnostics) {
		return
	}

	if err := r.client.ProjectOrganization.Add(
		ctx,
		plan.Project.ValueString(),
		plan.GroupID.ValueString(),
		plan.Role.ValueString(),
	); err != nil {
		resp.Diagnostics = util.DiagErrorCreatingResource(resp.Diagnostics, r, err)

		return
	}

	plan.ID = types.StringValue(util.ComposeID(plan.Project.ValueString(), plan.GroupID.ValueString()))

	err := r.fillModel(ctx, &plan)
	if err != nil {
		resp.Diagnostics = util.DiagErrorCreatingResource(resp.Diagnostics, r, err)

		return
	}

	if !util.ModelToPlanState(ctx, plan, &resp.State, &resp.Diagnostics) {
		return
	}
}

// Read reads the existing state of the resource.
func (r *organizationGroupProjectResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state organizationGroupProjectResourceModel

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

// Update updates an organization group project resource.
func (r *organizationGroupProjectResource) Update(
	_ context.Context,
	_ resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	resp.Diagnostics = util.DiagErrorUpdatingResourceNotSupported(resp.Diagnostics, r)
}

// Delete deletes an organization group project relation resource.
func (r *organizationGroupProjectResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var plan organizationGroupProjectResourceModel

	if !util.PlanStateToModel(ctx, &req.State, &plan, &resp.Diagnostics) {
		return
	}

	if err := r.client.ProjectOrganization.Delete(
		ctx,
		plan.Project.ValueString(),
		plan.GroupID.ValueString(),
	); err != nil {
		resp.Diagnostics = util.DiagErrorDeletingResource(resp.Diagnostics, r, err)

		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *organizationGroupProjectResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	util.UnpackCompoundID(ctx, req, resp, "project", "group_id")
}
