package billing

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationbilling"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/diagnostics"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	providertypes "github.com/aiven/terraform-provider-aiven/internal/plugin/types"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/validation"
)

var (
	_ resource.Resource                = &organizationBillingGroupResource{}
	_ resource.ResourceWithConfigure   = &organizationBillingGroupResource{}
	_ resource.ResourceWithImportState = &organizationBillingGroupResource{}
)

// NewOrganizationBillingGroupResource is a constructor for the organization billing group resource.
func NewOrganizationBillingGroupResource() resource.Resource {
	return &organizationBillingGroupResource{}
}

// organizationBillingGroupResource is the organization billing group resource implementation.
type organizationBillingGroupResource struct {
	// client is the instance of the Aiven client to use.
	client avngen.Client

	// diag is the diagnostics helper.
	diag *diagnostics.DiagnosticsHelper
}

// organizationBillingGroupResourceModel is the model for the organization billing group resource.
type organizationBillingGroupResourceModel struct {
	OrganizationBillingGroupModel

	// Timeouts is the configuration for resource-specific timeouts
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

// TypeName returns the resource type name.
func (r *organizationBillingGroupResource) TypeName() string {
	return "organization_billing_group_resource"
}

// Metadata returns the metadata for the organization billing group resource.
func (r *organizationBillingGroupResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_organization_billing_group"
}

// Schema defines the schema for the organization billing group resource.
func (r *organizationBillingGroupResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resourceSchema := ResourceSchema()

	// Add timeouts attribute
	resourceSchema["timeouts"] = timeouts.Attributes(ctx, timeouts.Opts{
		Create: true,
		Read:   true,
		Update: true,
		Delete: true,
	})

	resp.Schema = schema.Schema{
		Description: "Creates and manages an organization billing group.",
		Attributes:  resourceSchema,
	}
}

// Configure sets up the organization billing group resource.
func (r *organizationBillingGroupResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
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

// validateRequiredFields validates that all required fields are set.
func validateRequiredFields(
	ctx context.Context,
	model *organizationBillingGroupResourceModel,
	diags *diag.Diagnostics,
	diagHelper *diagnostics.DiagnosticsHelper,
) {
	validation.ValidateRequiredStringField(model.OrganizationID, "organization_id", diags, diagHelper)
	validation.ValidateRequiredStringField(model.BillingAddressID, "billing_address_id", diags, diagHelper)
	validation.ValidateRequiredListField(ctx, model.BillingContactEmails, "billing_contact_emails", diags, diagHelper)
	validation.ValidateRequiredListField(ctx, model.BillingEmails, "billing_emails", diags, diagHelper)
	validation.ValidateRequiredStringField(model.BillingGroupName, "billing_group_name", diags, diagHelper)
	validation.ValidateRequiredStringField(model.ShippingAddressID, "shipping_address_id", diags, diagHelper)
}

// getEmailList converts a types.List to []organizationbilling.BillingContactEmailIn or []organizationbilling.BillingEmailIn
func getEmailList(ctx context.Context, list types.List, isContactEmail bool) (interface{}, error) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}

	var strEmails []string
	diags := list.ElementsAs(ctx, &strEmails, false)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to get email list: %v", diags)
	}

	if isContactEmail {
		emails := make([]organizationbilling.BillingContactEmailIn, len(strEmails))
		for i, email := range strEmails {
			emails[i] = organizationbilling.BillingContactEmailIn{Email: email}
		}
		return emails, nil
	}

	emails := make([]organizationbilling.BillingEmailIn, len(strEmails))
	for i, email := range strEmails {
		emails[i] = organizationbilling.BillingEmailIn{Email: email}
	}
	return emails, nil
}

// getEmailStrings extracts email strings from BillingContactEmailOut or BillingEmailOut slices
func getEmailStrings(emails interface{}) []string {
	switch e := emails.(type) {
	case []organizationbilling.BillingContactEmailOut:
		result := make([]string, len(e))
		for i, email := range e {
			result[i] = email.Email
		}
		return result
	case []organizationbilling.BillingEmailOut:
		result := make([]string, len(e))
		for i, email := range e {
			result[i] = email.Email
		}
		return result
	default:
		return nil
	}
}

// Create creates an organization billing group resource.
func (r *organizationBillingGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationBillingGroupResourceModel

	// Get plan values
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate required fields
	validateRequiredFields(ctx, &plan, &resp.Diagnostics, r.diag)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get email lists
	contactEmails, err := getEmailList(ctx, plan.BillingContactEmails, true)
	if err != nil {
		resp.Diagnostics.AddError("Error getting billing contact emails", err.Error())
		return
	}
	billingEmails, err := getEmailList(ctx, plan.BillingEmails, false)
	if err != nil {
		resp.Diagnostics.AddError("Error getting billing emails", err.Error())
		return
	}

	// Convert currency to BillingCurrencyType
	var billingCurrency organizationbilling.BillingCurrencyType
	if !plan.BillingCurrency.IsNull() {
		billingCurrency = organizationbilling.BillingCurrencyType(plan.BillingCurrency.ValueString())
	}

	// Prepare the request body
	createReq := &organizationbilling.OrganizationBillingGroupCreateIn{
		BillingAddressId:     plan.BillingAddressID.ValueString(),
		BillingContactEmails: contactEmails.([]organizationbilling.BillingContactEmailIn),
		BillingCurrency:      billingCurrency,
		BillingEmails:        billingEmails.([]organizationbilling.BillingEmailIn),
		BillingGroupName:     plan.BillingGroupName.ValueString(),
		CustomInvoiceText:    plan.CustomInvoiceText.ValueStringPointer(),
		PaymentMethodId:      plan.PaymentMethodID.ValueStringPointer(),
		ShippingAddressId:    plan.ShippingAddressID.ValueString(),
		VatId:                plan.VATID.ValueStringPointer(),
	}

	// Create the billing group
	billingGroup, err := r.client.OrganizationBillingGroupCreate(ctx, plan.OrganizationID.ValueString(), createReq)
	if err != nil {
		r.diag.AddError(&resp.Diagnostics, "creating", err)
		return
	}

	// Set state
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", plan.OrganizationID.ValueString(), billingGroup.BillingGroupId))
	plan.BillingGroupID = types.StringValue(billingGroup.BillingGroupId)

	// Convert email lists to types.List
	contactEmailsList, diags := types.ListValueFrom(ctx, types.StringType, getEmailStrings(billingGroup.BillingContactEmails))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.BillingContactEmails = contactEmailsList

	billingEmailsList, diags := types.ListValueFrom(ctx, types.StringType, getEmailStrings(billingGroup.BillingEmails))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.BillingEmails = billingEmailsList

	plan.BillingCurrency = types.StringValue(string(billingGroup.BillingCurrency))
	plan.CustomInvoiceText = types.StringPointerValue(billingGroup.CustomInvoiceText)
	plan.VATID = types.StringPointerValue(billingGroup.VatId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads an organization billing group resource.
func (r *organizationBillingGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationBillingGroupResourceModel

	// Get state values
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the billing group
	billingGroup, err := r.client.OrganizationBillingGroupGet(ctx, state.OrganizationID.ValueString(), state.BillingGroupID.ValueString())
	if err != nil {
		r.diag.AddError(&resp.Diagnostics, "reading", err)
		return
	}

	// Set state
	state.ID = types.StringValue(fmt.Sprintf("%s/%s", state.OrganizationID.ValueString(), billingGroup.BillingGroupId))
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

// Update updates an organization billing group resource.
func (r *organizationBillingGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and state
	var plan, state organizationBillingGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate required fields
	validateRequiredFields(ctx, &plan, &resp.Diagnostics, r.diag)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get email lists
	contactEmails, err := getEmailList(ctx, plan.BillingContactEmails, true)
	if err != nil {
		resp.Diagnostics.AddError("Error getting billing contact emails", err.Error())
		return
	}
	billingEmails, err := getEmailList(ctx, plan.BillingEmails, false)
	if err != nil {
		resp.Diagnostics.AddError("Error getting billing emails", err.Error())
		return
	}

	// Convert currency to BillingCurrencyType
	var billingCurrency organizationbilling.BillingCurrencyType
	if !plan.BillingCurrency.IsNull() {
		billingCurrency = organizationbilling.BillingCurrencyType(plan.BillingCurrency.ValueString())
	}

	// Convert strings to pointers where needed
	billingAddressID := plan.BillingAddressID.ValueString()
	billingGroupName := plan.BillingGroupName.ValueString()
	paymentMethodID := plan.PaymentMethodID.ValueString()
	shippingAddressID := plan.ShippingAddressID.ValueString()

	// Prepare the request body
	updateReq := &organizationbilling.OrganizationBillingGroupUpdateIn{
		BillingAddressId:     &billingAddressID,
		BillingContactEmails: contactEmails.([]organizationbilling.BillingContactEmailIn),
		BillingCurrency:      billingCurrency,
		BillingEmails:        billingEmails.([]organizationbilling.BillingEmailIn),
		BillingGroupName:     billingGroupName,
		CustomInvoiceText:    plan.CustomInvoiceText.ValueStringPointer(),
		PaymentMethodId:      &paymentMethodID,
		ShippingAddressId:    &shippingAddressID,
		VatId:                plan.VATID.ValueStringPointer(),
	}

	// Update the billing group
	billingGroup, err := r.client.OrganizationBillingGroupUpdate(ctx, plan.OrganizationID.ValueString(), state.BillingGroupID.ValueString(), updateReq)
	if err != nil {
		r.diag.AddError(&resp.Diagnostics, "updating", err)
		return
	}

	// Set state
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", plan.OrganizationID.ValueString(), billingGroup.BillingGroupId))
	plan.BillingGroupID = types.StringValue(billingGroup.BillingGroupId)

	// Convert email lists to types.List
	contactEmailsList, diags := types.ListValueFrom(ctx, types.StringType, getEmailStrings(billingGroup.BillingContactEmails))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.BillingContactEmails = contactEmailsList

	billingEmailsList, diags := types.ListValueFrom(ctx, types.StringType, getEmailStrings(billingGroup.BillingEmails))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.BillingEmails = billingEmailsList

	plan.BillingCurrency = types.StringValue(string(billingGroup.BillingCurrency))
	plan.CustomInvoiceText = types.StringPointerValue(billingGroup.CustomInvoiceText)
	plan.VATID = types.StringPointerValue(billingGroup.VatId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes an organization billing group resource.
func (r *organizationBillingGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationBillingGroupResourceModel

	// Get state values
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the billing group
	err := r.client.OrganizationBillingGroupDelete(ctx, state.OrganizationID.ValueString(), state.BillingGroupID.ValueString())
	if err != nil {
		r.diag.AddError(&resp.Diagnostics, "deleting", err)
		return
	}
}

// ImportState imports an organization billing group resource.
func (r *organizationBillingGroupResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Check if import ID is valid
	parts, err := validation.ValidateImportID(req.ID, "organization_id/billing_group_id")
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}

	// Set the import attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("billing_group_id"), parts[1])...)
}
