package address

import (
	"context"
	"fmt"

	"github.com/aiven/go-client-codegen/handler/organization"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/diagnostics"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	providertypes "github.com/aiven/terraform-provider-aiven/internal/plugin/types"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/validation"
)

var (
	_ datasource.DataSource              = &organizationAddressDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationAddressDataSource{}
)

// NewOrganizationAddressDataSource is a constructor for the organization address data source.
func NewOrganizationAddressDataSource() datasource.DataSource {
	return &organizationAddressDataSource{}
}

// organizationAddressDataSource is the organization address data source implementation.
type organizationAddressDataSource struct {
	// client is the instance of the Aiven client to use.
	client organization.Handler

	// diag is the diagnostics helper.
	diag *diagnostics.DiagnosticsHelper
}

// organizationAddressDataSourceModel is the model for the organization address data source.
type organizationAddressDataSourceModel struct {
	OrganizationAddressModel
}

// TypeName returns the data source type name.
func (r *organizationAddressDataSource) TypeName() string {
	return "organization_address_data_source"
}

// Metadata returns the metadata for the organization address data source.
func (r *organizationAddressDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_organization_address"
}

// Schema defines the schema for the organization address data source.
func (r *organizationAddressDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Gets information about an organization address.",
		Attributes:  util.ResourceSchemaToDataSourceSchema(ResourceSchema(), []string{}, []string{"organization_id", "address_id"}),
	}
}

// Configure sets up the organization address data source.
func (r *organizationAddressDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	p, ok := req.ProviderData.(providertypes.AivenClientProvider)
	if !ok {
		resp.Diagnostics.AddError(
			errmsg.SummaryUnexpectedProviderDataType,
			fmt.Sprintf(errmsg.DetailUnexpectedProviderDataType, req.ProviderData),
		)
		return
	}

	r.client = p.GetGenClient()
	r.diag = diagnostics.NewDiagnosticsHelper(r.TypeName())
}

// validateDataSourceRequiredFields validates that all required fields are set.
func validateDataSourceRequiredFields(
	model *organizationAddressDataSourceModel,
	diags *diag.Diagnostics,
	diagHelper *diagnostics.DiagnosticsHelper,
) {
	// Validate organization_id only - address_id validation is handled by the schema
	validation.ValidateRequiredStringField(model.OrganizationID, "organization_id", diags, diagHelper)
}

// ConfigValidators returns the configuration validators for the organization address data source.
func (r *organizationAddressDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.RequiredTogether(
			path.MatchRoot("organization_id"),
			path.MatchRoot("address_id"),
		),
	}
}

// Read reads an organization address data source.
func (r *organizationAddressDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationAddressDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate required fields
	validateDataSourceRequiredFields(&config, &resp.Diagnostics, r.diag)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the address
	address, err := r.client.OrganizationAddressGet(ctx, config.OrganizationID.ValueString(), config.AddressID.ValueString())
	if err != nil {
		r.diag.AddError(&resp.Diagnostics, "reading", err)
		return
	}

	// Convert address lines to types.List
	addressLines, diags := types.ListValueFrom(ctx, types.StringType, address.AddressLines)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set all fields
	state := organizationAddressDataSourceModel{}
	state.ID = types.StringValue(fmt.Sprintf("%s/%s", config.OrganizationID.ValueString(), address.AddressId))
	state.OrganizationID = config.OrganizationID
	state.AddressID = types.StringValue(address.AddressId)
	state.AddressLines = addressLines
	state.City = types.StringPointerValue(address.City)
	state.CompanyName = types.StringPointerValue(address.CompanyName)
	state.CountryCode = types.StringValue(address.CountryCode)
	state.State = types.StringPointerValue(address.State)
	state.ZipCode = types.StringPointerValue(address.ZipCode)
	state.CreateTime = types.StringValue(address.CreateTime.String())
	state.UpdateTime = types.StringValue(address.UpdateTime.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
