package organization

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

var (
	_ datasource.DataSource              = &organizationDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationDataSource{}

	_ util.TypeNameable = &organizationDataSource{}
)

// NewOrganizationDataSource is a constructor for the organization data source.
func NewOrganizationDataSource() datasource.DataSource {
	return &organizationDataSource{}
}

// organizationDataSource is the organization data source implementation.
type organizationDataSource struct {
	// client is the instance of the Aiven client to use.
	client *aiven.Client

	// typeName is the name of the data source type.
	typeName string
}

// organizationDataSourceModel is the model for the organization data source.
type organizationDataSourceModel struct {
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
}

// Metadata returns the metadata for the organization data source.
func (r *organizationDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_organization"

	r.typeName = resp.TypeName
}

// TypeName returns the data source type name for the organization data source.
func (r *organizationDataSource) TypeName() string {
	return r.typeName
}

// Schema defines the schema for the organization data source.
func (r *organizationDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an organization from Aiven.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier of the organization.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the organization.",
				Optional:    true,
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
	}
}

// Configure sets up the organization data source.
func (r *organizationDataSource) Configure(
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

// ConfigValidators returns the configuration validators for the organization data source.
func (r *organizationDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	validatable := path.Expressions{
		path.MatchRoot("id"),
		path.MatchRoot("name"),
	}

	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(validatable...),
		datasourcevalidator.AtLeastOneOf(validatable...),
	}
}

// fillModel fills the organization data source model from the Aiven API.
func (r *organizationDataSource) fillModel(ctx context.Context, model *organizationDataSourceModel) (err error) {
	normalizedID, err := schemautil.NormalizeOrganizationID(ctx, r.client, model.ID.ValueString())
	if err != nil {
		return
	}

	account, err := r.client.Accounts.Get(ctx, normalizedID)
	if err != nil {
		return
	}

	model.Name = types.StringValue(account.Account.Name)

	model.TenantID = types.StringValue(account.Account.TenantId)

	model.CreateTime = types.StringValue(account.Account.CreateTime.String())

	model.UpdateTime = types.StringValue(account.Account.UpdateTime.String())

	return
}

// Read reads an organization data source.
func (r *organizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state organizationDataSourceModel

	if !util.ConfigToModel(ctx, &req.Config, &state, &resp.Diagnostics) {
		return
	}

	if state.ID.IsNull() {
		list, err := r.client.Accounts.List(ctx)
		if err != nil {
			resp.Diagnostics = util.DiagErrorReadingDataSource(resp.Diagnostics, r, err)

			return
		}

		var found bool

		for _, account := range list.Accounts {
			if account.Name == state.Name.ValueString() {
				if found {
					resp.Diagnostics = util.DiagDuplicateFoundByName(resp.Diagnostics, state.Name.ValueString())

					return
				}

				state.ID = types.StringValue(account.OrganizationId)

				found = true
			}
		}
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
