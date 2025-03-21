package billing

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
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
	_ datasource.DataSource              = &organizationBillingGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationBillingGroupDataSource{}
)

// NewOrganizationBillingGroupDataSource is a constructor for the organization billing group data source.
func NewOrganizationBillingGroupDataSource() datasource.DataSource {
	return &organizationBillingGroupDataSource{}
}

// organizationBillingGroupDataSource is the organization billing group data source implementation.
type organizationBillingGroupDataSource struct {
	// client is the instance of the Aiven client to use.
	client avngen.Client

	// diag is the diagnostics helper.
	diag *diagnostics.DiagnosticsHelper
}

// organizationBillingGroupDataSourceModel is the model for the organization billing group data source.
type organizationBillingGroupDataSourceModel struct {
	OrganizationBillingGroupModel
}

// TypeName returns the data source type name.
func (r *organizationBillingGroupDataSource) TypeName() string {
	return "organization_billing_group_data_source"
}

// Metadata returns the metadata for the organization billing group data source.
func (r *organizationBillingGroupDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_organization_billing_group"
}

// Schema defines the schema for the organization billing group data source.
func (r *organizationBillingGroupDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Gets information about a billing group.",
		Attributes:  util.ResourceSchemaToDataSourceSchema(ResourceSchema(), []string{}, []string{"organization_id", "billing_group_id"}),
	}
}

// Configure sets up the organization billing group data source.
func (r *organizationBillingGroupDataSource) Configure(
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
	model *organizationBillingGroupDataSourceModel,
	diags *diag.Diagnostics,
	diagHelper *diagnostics.DiagnosticsHelper,
) {
	// Validate organization_id only - billing_group_id validation is handled by the schema
	validation.ValidateRequiredStringField(model.OrganizationID, "organization_id", diags, diagHelper)
}

// ConfigValidators returns the configuration validators for the organization billing group data source.
func (r *organizationBillingGroupDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.RequiredTogether(
			path.MatchRoot("organization_id"),
			path.MatchRoot("billing_group_id"),
		),
	}
}

// Read reads an organization billing group data source.
func (r *organizationBillingGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationBillingGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate required fields
	validateDataSourceRequiredFields(&config, &resp.Diagnostics, r.diag)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the billing group
	billingGroup, err := r.client.OrganizationBillingGroupGet(ctx, config.OrganizationID.ValueString(), config.BillingGroupID.ValueString())
	if err != nil {
		r.diag.AddError(&resp.Diagnostics, "reading", err)
		return
	}

	// Set all fields
	state := organizationBillingGroupDataSourceModel{}
	state.ID = types.StringValue(fmt.Sprintf("%s/%s", config.OrganizationID.ValueString(), billingGroup.BillingGroupId))
	state.OrganizationID = config.OrganizationID
	state.BillingGroupID = types.StringValue(billingGroup.BillingGroupId)
	state.BillingAddressID = types.StringValue(billingGroup.BillingAddressId)
	state.ShippingAddressID = types.StringValue(billingGroup.ShippingAddressId)
	state.PaymentMethodID = types.StringPointerValue(billingGroup.PaymentMethodId)
	state.BillingGroupName = types.StringValue(billingGroup.BillingGroupName)

	// Convert email lists to types.List
	contactEmailsList, diags := types.ListValueFrom(ctx, types.StringType, getEmailStrings(billingGroup.BillingContactEmails))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.BillingContactEmails = contactEmailsList

	billingEmailsList, diags := types.ListValueFrom(ctx, types.StringType, getEmailStrings(billingGroup.BillingEmails))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.BillingEmails = billingEmailsList

	state.BillingCurrency = types.StringValue(string(billingGroup.BillingCurrency))
	state.CustomInvoiceText = types.StringPointerValue(billingGroup.CustomInvoiceText)
	state.VATID = types.StringPointerValue(billingGroup.VatId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
