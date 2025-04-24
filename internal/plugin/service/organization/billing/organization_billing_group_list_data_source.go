package billing

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/diagnostics"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	providertypes "github.com/aiven/terraform-provider-aiven/internal/plugin/types"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/validation"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var (
	_ datasource.DataSource              = &organizationBillingGroupListDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationBillingGroupListDataSource{}
)

// NewOrganizationBillingGroupListDataSource is a constructor for the organization billing group list data source.
func NewOrganizationBillingGroupListDataSource() datasource.DataSource {
	return &organizationBillingGroupListDataSource{}
}

// organizationBillingGroupListDataSource is the organization billing group list data source implementation.
type organizationBillingGroupListDataSource struct {
	// client is the instance of the Aiven client to use.
	client avngen.Client

	// diag is the diagnostics helper.
	diag *diagnostics.DiagnosticsHelper
}

// organizationBillingGroupListDataSourceModel is the model for the organization billing group list data source.
type organizationBillingGroupListDataSourceModel struct {
	// ID is a composite of organization_id
	ID types.String `tfsdk:"id"`

	// OrganizationID is the identifier of the organization
	OrganizationID types.String `tfsdk:"organization_id"`

	// BillingGroups is the list of billing groups
	BillingGroups []OrganizationBillingGroupModel `tfsdk:"billing_groups"`
}

// TypeName returns the data source type name.
func (r *organizationBillingGroupListDataSource) TypeName() string {
	return "organization_billing_group_list_data_source"
}

// Metadata returns the metadata for the organization billing group list data source.
func (r *organizationBillingGroupListDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_organization_billing_group_list"
}

// Schema defines the schema for the organization billing group list data source.
func (r *organizationBillingGroupListDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	// Convert resource schema to datasource schema
	dataSourceSchema := util.ResourceSchemaToDataSourceSchema(ResourceSchema(), []string{}, []string{})

	resp.Schema = schema.Schema{
		Description: userconfig.Desc("Lists billing groups for an organization.").AvailabilityType(userconfig.Beta).Build(),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource ID, a composite of organization_id.",
				Computed:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "ID of the organization.",
				Required:    true,
			},
			"billing_groups": schema.ListNestedAttribute{
				Description: "List of billing groups.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: dataSourceSchema,
				},
			},
		},
	}
}

// Configure sets up the organization billing group list data source.
func (r *organizationBillingGroupListDataSource) Configure(
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

// validateListDataSourceRequiredFields validates that all required fields are set.
func validateListDataSourceRequiredFields(
	model *organizationBillingGroupListDataSourceModel,
	diags *diag.Diagnostics,
	diagHelper *diagnostics.DiagnosticsHelper,
) {
	// Validate organization_id
	validation.ValidateRequiredStringField(model.OrganizationID, "organization_id", diags, diagHelper)
}

// Read reads an organization billing group list data source.
func (r *organizationBillingGroupListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationBillingGroupListDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate required fields
	validateListDataSourceRequiredFields(&config, &resp.Diagnostics, r.diag)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the billing groups
	billingGroups, err := r.client.OrganizationBillingGroupList(ctx, config.OrganizationID.ValueString())
	if err != nil {
		r.diag.AddError(&resp.Diagnostics, "reading", err)
		return
	}

	// Set state
	state := organizationBillingGroupListDataSourceModel{
		ID:             types.StringValue(config.OrganizationID.ValueString()),
		OrganizationID: config.OrganizationID,
		BillingGroups:  make([]OrganizationBillingGroupModel, 0, len(billingGroups)),
	}

	// Convert billing groups
	for _, bg := range billingGroups {
		// Convert email lists to types.Set
		contactEmailsList, diags := types.SetValueFrom(ctx, types.StringType, getEmailStrings(bg.BillingContactEmails))
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		billingEmailsList, diags := types.SetValueFrom(ctx, types.StringType, getEmailStrings(bg.BillingEmails))
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Create model
		model := OrganizationBillingGroupModel{
			ID:                   types.StringValue(fmt.Sprintf("%s/%s", config.OrganizationID.ValueString(), bg.BillingGroupId)),
			OrganizationID:       config.OrganizationID,
			BillingGroupID:       types.StringValue(bg.BillingGroupId),
			BillingAddressID:     types.StringValue(bg.BillingAddressId),
			BillingContactEmails: contactEmailsList,
			BillingCurrency:      types.StringValue(string(bg.BillingCurrency)),
			BillingEmails:        billingEmailsList,
			BillingGroupName:     types.StringValue(bg.BillingGroupName),
			CustomInvoiceText:    types.StringPointerValue(bg.CustomInvoiceText),
			PaymentMethodID:      types.StringPointerValue(bg.PaymentMethodId),
			ShippingAddressID:    types.StringValue(bg.ShippingAddressId),
			VATID:                types.StringPointerValue(bg.VatId),
		}

		state.BillingGroups = append(state.BillingGroups, model)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
