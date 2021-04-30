# Billing Group Resource

The Billing Group resource allows the creation and management of Aiven Billing Groups and association with the Projects.

## Example Usage

```hcl
resource "aiven_billing_group" "bybg1" {
  name = "bybg1"
  billing_currency = "USD"
  vat_id = "123ABC"
}

resource "aiven_project" "pr1" {
  project = "pr1"
  billing_group = aiven_billing_group.bybg1.id
}
```

## Argument Reference

* `name` - (Required) Billing Group name

* `card_id` - (Optional) Credit card id

* `vat_id` - (Optional) VAT id

* `account_id` - (Optional) Account id

* `billing_currency` - (Optional) Billing currency

* `billing_extra_text` - (Optional) Billing extra text

* `billing_emails` - (Optional) Billing contact emails

* `company` - (Optional) Company name

* `address_lines` - (Optional) Address lines

* `country_code` - (Optional) Country code

* `city` - (Optional) City

* `zip_code` - (Optional) Zip Code

* `state` - (Optional) State

Aiven ID format when importing existing resource: `<billing_group_id>`