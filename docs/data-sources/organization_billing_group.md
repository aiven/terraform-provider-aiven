---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_organization_billing_group Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about a billing group.
---

# aiven_organization_billing_group (Data Source)

Gets information about a billing group.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `billing_group_id` (String) ID of the billing group.
- `organization_id` (String) ID of the organization.

### Read-Only

- `billing_address_id` (String) ID of the billing address.
- `billing_contact_emails` (List of String) List of billing contact emails.
- `billing_currency` (String) Billing currency.
- `billing_emails` (List of String) List of billing emails.
- `billing_group_name` (String) Name of the billing group.
- `custom_invoice_text` (String) Custom invoice text.
- `id` (String) Resource ID, a composite of organization_id and billing_group_id.
- `payment_method_id` (String) ID of the payment method.
- `shipping_address_id` (String) ID of the shipping address.
- `vat_id` (String) VAT ID.
