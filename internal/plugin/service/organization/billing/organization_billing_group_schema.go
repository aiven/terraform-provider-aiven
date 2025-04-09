package billing

import (
	"github.com/aiven/go-client-codegen/handler/account"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// OrganizationBillingGroupModel is the common model for both resource and data source
type OrganizationBillingGroupModel struct {
	// ID is a composite of organization_id and billing_group_id
	ID types.String `tfsdk:"id"`

	// OrganizationID is the identifier of the organization
	OrganizationID types.String `tfsdk:"organization_id"`

	// BillingGroupID is the identifier of the billing group
	BillingGroupID types.String `tfsdk:"billing_group_id"`

	// BillingAddressID is the identifier of the billing address
	BillingAddressID types.String `tfsdk:"billing_address_id"`

	// BillingContactEmails is the list of billing contact emails
	BillingContactEmails types.Set `tfsdk:"billing_contact_emails"`

	// BillingCurrency is the billing currency
	BillingCurrency types.String `tfsdk:"billing_currency"`

	// BillingEmails is the list of billing emails
	BillingEmails types.Set `tfsdk:"billing_emails"`

	// BillingGroupName is the name of the billing group
	BillingGroupName types.String `tfsdk:"billing_group_name"`

	// CustomInvoiceText is the custom invoice text
	CustomInvoiceText types.String `tfsdk:"custom_invoice_text"`

	// PaymentMethodID is the identifier of the payment method
	PaymentMethodID types.String `tfsdk:"payment_method_id"`

	// ShippingAddressID is the identifier of the shipping address
	ShippingAddressID types.String `tfsdk:"shipping_address_id"`

	// VATID is the VAT ID
	VATID types.String `tfsdk:"vat_id"`
}

// ResourceSchema returns the schema for the organization billing group resource
func ResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Resource ID, a composite of organization_id and billing_group_id.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"organization_id": schema.StringAttribute{
			Description: "ID of the organization.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"billing_group_id": schema.StringAttribute{
			Description: "ID of the billing group.",
			Computed:    true,
		},
		"billing_address_id": schema.StringAttribute{
			Description: "ID of the billing address.",
			Required:    true,
		},
		"billing_contact_emails": schema.SetAttribute{
			Description: "List of billing contact emails.",
			Required:    true,
			ElementType: types.StringType,
		},
		"billing_currency": schema.StringAttribute{
			Description: "Billing currency.",
			Optional:    true,
			Validators: []validator.String{
				stringvalidator.OneOf(account.BillingCurrencyTypeChoices()...),
			},
		},
		"billing_emails": schema.SetAttribute{
			Description: "List of billing emails.",
			Required:    true,
			ElementType: types.StringType,
		},
		"billing_group_name": schema.StringAttribute{
			Description: "Name of the billing group.",
			Required:    true,
		},
		"custom_invoice_text": schema.StringAttribute{
			Description: "Custom invoice text.",
			Optional:    true,
		},
		"payment_method_id": schema.StringAttribute{
			Description: "ID of the payment method.",
			Required:    true,
		},
		"shipping_address_id": schema.StringAttribute{
			Description: "ID of the shipping address.",
			Required:    true,
		},
		"vat_id": schema.StringAttribute{
			Description: "VAT ID.",
			Optional:    true,
		},
	}
}
