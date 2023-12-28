package organization

import (
	"context"
	"fmt"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

var (
	_ resource.Resource                = &organizationGroupProjectResource{}
	_ resource.ResourceWithConfigure   = &organizationGroupProjectResource{}
	_ resource.ResourceWithImportState = &organizationGroupProjectResource{}

	_ util.TypeNameable = &organizationGroupProjectResource{}
)

// NewOrganizationGroupProjectResource is a constructor for the organization resource.
func NewOrganizationGroupProjectResource() resource.Resource {
	return &organizationGroupProjectResource{}
}

// organizationGroupUserResource is the organization resource implementation.
type organizationGroupProjectResource struct {
	// client is the instance of the Aiven client to use.
	client *aiven.Client

	// typeName is the name of the resource type.
	typeName string
}

// organizationGroupProjectResourceModel is the model for the organization resource.
type organizationGroupProjectResourceModel struct {
	// Name is the name of the organization.
	Project types.String `tfsdk:"project"`
	// OrganizationID is the identifier of the organization group.
	OrganizationGroupID types.String `tfsdk:"group_id"`
	// Role is the role of the organization group project relation.
	Role types.String `tfsdk:"role"`
	// Timeouts is the configuration for resource-specific timeouts.
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

// Metadata returns the metadata for the organization resource.
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
		Description: "Creates and manages an organization group project relations in Aiven.",
		Attributes: map[string]schema.Attribute{
			"group_id": schema.StringAttribute{
				Description: "Organization group identifier of the organization group project relation.",
				Required:    true,
			},
			"project": schema.StringAttribute{
				Description: "Tenant identifier of the organization.",
				Required:    true,
			},
			"role": schema.StringAttribute{
				Description: "Role of the organization group project relation.",
				Required:    true,
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

// CustomizeDiff helps to customize the diff for the resource.
func (r *organizationGroupProjectResource) fillModel(
	ctx context.Context,
	m *organizationGroupProjectResourceModel,
) error {
	list, err := r.client.ProjectOrganization.List(
		ctx,
		m.Project.ValueString())
	if err != nil {
		return err
	}

	var isFound bool
	for _, project := range list {
		if project.OrganizationGroupID == m.OrganizationGroupID.ValueString() {
			isFound = true
			m.OrganizationGroupID = types.StringValue(project.OrganizationGroupID)
			m.Role = types.StringValue(project.Role)
		}
	}

	if !isFound {
		return fmt.Errorf("organization group project relation not found, organization group id: %s, project: %s",
			m.OrganizationGroupID.ValueString(), m.Project.ValueString())
	}

	// There is not API endpoint to get the permission of the organization group project relation.

	return nil
}

// Diff helps to differentiate desired from the existing state of the resource.
func (r *organizationGroupProjectResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan organizationGroupProjectResourceModel
	if !util.PlanStateToModel(ctx, &req.Plan, &plan, &resp.Diagnostics) {
		return
	}

	err := r.client.ProjectOrganization.Add(
		ctx,
		plan.Project.ValueString(),
		plan.OrganizationGroupID.ValueString(),
		plan.Role.ValueString())
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
func (r *organizationGroupProjectResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var plan organizationGroupProjectResourceModel

	if !util.PlanStateToModel(ctx, &req.State, &plan, &resp.Diagnostics) {
		return
	}

	err := r.client.ProjectOrganization.Delete(
		ctx,
		plan.Project.ValueString(),
		plan.OrganizationGroupID.ValueString())
	if err != nil {
		resp.Diagnostics = util.DiagErrorDeletingResource(resp.Diagnostics, r, err)

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

// ImportState imports an existing resource into Terraform.
func (r *organizationGroupProjectResource) ImportState(
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

// Update updates an organization group project resource.
func (r *organizationGroupProjectResource) Update(
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
