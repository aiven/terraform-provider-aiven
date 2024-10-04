package externalidentity

import (
	"context"
	"fmt"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var (
	_ datasource.DataSource              = &externalIdentityDataSource{}
	_ datasource.DataSourceWithConfigure = &externalIdentityDataSource{}
	_ util.TypeNameable                  = &externalIdentityDataSource{}
)

func NewExternalIdentityDataSource() datasource.DataSource {
	return &externalIdentityDataSource{}
}

type externalIdentityDataSource struct {
	client   *aiven.Client
	typeName string
}

// externalIdentityDataSourceModel is the model for the external_identity data source.
type externalIdentityDataSourceModel struct {
	OrganizationID      types.String `tfsdk:"organization_id"`
	InternalUserID      types.String `tfsdk:"internal_user_id"`
	ExternalUserID      types.String `tfsdk:"external_user_id"`
	ExternalServiceName types.String `tfsdk:"external_service_name"`
}

// Metadata returns the metadata for the external_identity data source.
func (r *externalIdentityDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_external_identity"
	r.typeName = resp.TypeName
}

// TypeName returns the data source type name for the external_identity data source.
func (r *externalIdentityDataSource) TypeName() string {
	return r.typeName
}

// Schema defines the schema for the external_identity data source.
func (r *externalIdentityDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: userconfig.Desc("Maps an external service user to an Aiven user.").AvailabilityType(userconfig.Beta).Build(),
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.StringAttribute{
				Description: "The ID of the Aiven organization that the user is part of.",
				Required:    true,
			},
			"internal_user_id": schema.StringAttribute{
				Description: "The Aiven user ID.",
				Required:    true,
			},
			"external_user_id": schema.StringAttribute{
				Description: "The user's ID on the external service.",
				Required:    true,
			},
			"external_service_name": schema.StringAttribute{
				Description: userconfig.Desc("The name of the external service.").PossibleValuesString("github").Build(),
				Required:    true,
			},
		},
	}
}

// Configure sets up the external_identity data source.
func (r *externalIdentityDataSource) Configure(
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

// Read reads an external_identity data source.
func (r *externalIdentityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state externalIdentityDataSourceModel

	if !util.ConfigToModel(ctx, &req.Config, &state, &resp.Diagnostics) {
		return
	}

	organizationID := state.OrganizationID.ValueString()
	internalUserID := state.InternalUserID.ValueString()
	externalUserID := state.ExternalUserID.ValueString()
	externalServiceName := state.ExternalServiceName.ValueString()

	responseData, err := r.client.OrganizationUser.List(ctx, organizationID)
	if err != nil {
		resp.Diagnostics = util.DiagErrorReadingDataSource(resp.Diagnostics, r, err)
		return
	}

	var user *aiven.OrganizationMemberInfo
	for _, member := range responseData.Users {
		if member.UserID == internalUserID {
			user = &member
			break
		}
	}

	if user == nil {
		err := fmt.Errorf("organization user %s not found in organization %s", internalUserID, organizationID)
		resp.Diagnostics = util.DiagErrorReadingDataSource(resp.Diagnostics, r, err)
		return
	}

	state.OrganizationID = types.StringValue(organizationID)
	state.InternalUserID = types.StringValue(internalUserID)
	state.ExternalUserID = types.StringValue(externalUserID)
	state.ExternalServiceName = types.StringValue(externalServiceName)

	util.ModelToPlanState(ctx, state, &resp.State, &resp.Diagnostics)
}
