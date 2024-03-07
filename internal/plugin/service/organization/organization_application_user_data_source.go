package organization

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var (
	_ datasource.DataSource              = &organizationApplicationUserDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationApplicationUserDataSource{}

	_ util.TypeNameable = &organizationApplicationUserDataSource{}
)

// NewOrganizationApplicationUserDataSource is a constructor for the organization application user data source.
func NewOrganizationApplicationUserDataSource() datasource.DataSource {
	return &organizationApplicationUserDataSource{}
}

// organizationApplicationUserDataSource is the organization application user data source implementation.
type organizationApplicationUserDataSource struct {
	// client is the instance of the Aiven client to use.
	client *aiven.Client

	// typeName is the name of the data source type.
	typeName string
}

// organizationApplicationUserDataSourceModel is the model for the organization application user data source.
type organizationApplicationUserDataSourceModel struct {
	// OrganizationID is the identifier of the organization the application user belongs to.
	OrganizationID types.String `tfsdk:"organization_id"`
	// UserID is the identifier of the organization application user.
	UserID types.String `tfsdk:"user_id"`
	// Name is the name of the organization application user.
	Name types.String `tfsdk:"name"`
	// Email is the email of the organization application user.
	Email types.String `tfsdk:"email"`
}

// Metadata returns the metadata for the organization application user data source.
func (r *organizationApplicationUserDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_organization_application_user"

	r.typeName = resp.TypeName
}

// TypeName returns the data source type name for the organization application user data source.
func (r *organizationApplicationUserDataSource) TypeName() string {
	return r.typeName
}

// Schema defines the schema for the organization application user data source.
func (r *organizationApplicationUserDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: userconfig.
			Desc(
				"Gets information about an application user.",
			).
			MarkAsDataSource().
			AvailabilityType(userconfig.Limited).
			Build(),
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.StringAttribute{
				Description: "The ID of the organization the application user belongs to.",
				Required:    true,
			},
			"user_id": schema.StringAttribute{
				Description: "The ID of the application user.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the application user.",
				Computed:    true,
			},
			"email": schema.StringAttribute{
				Description: "The auto-generated email address of the application user.",
				Computed:    true,
			},
		},
	}
}

// Configure sets up the organization application user data source.
func (r *organizationApplicationUserDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
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

// fillModel fills the organization application user data source model from the Aiven API.
func (r *organizationApplicationUserDataSource) fillModel(
	ctx context.Context,
	model *organizationApplicationUserDataSourceModel,
) (err error) {
	appUsers, err := r.client.OrganizationApplicationUserHandler.List(ctx, model.OrganizationID.ValueString())
	if err != nil {
		return
	}

	var appUser *aiven.ApplicationUserInfo

	for _, u := range appUsers.Users {
		if u.UserID == model.UserID.ValueString() {
			appUser = &u
			break
		}
	}

	model.Name = types.StringValue(appUser.Name)

	model.Email = types.StringValue(appUser.UserEmail)

	return
}

// Read reads an organization application user data source.
func (r *organizationApplicationUserDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state organizationApplicationUserDataSourceModel

	if !util.ConfigToModel(ctx, &req.Config, &state, &resp.Diagnostics) {
		return
	}

	err := r.fillModel(ctx, &state)
	if err != nil {
		resp.Diagnostics = util.DiagErrorReadingDataSource(resp.Diagnostics, r, err)

		return
	}

	if !util.ModelToPlanState(ctx, state, &resp.State, &resp.Diagnostics) {
		return
	}
}
