---
page_title: "aiven_cmk_accessor_azure Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets the Azure CMK accessor for an Aiven project. The accessor is used to authenticate Aiven to your Azure Key Vault.
---

# aiven_cmk_accessor_azure (Data Source)

Gets the Azure CMK accessor for an Aiven project. The accessor is used to authenticate Aiven to your Azure Key Vault.

## Example Usage

```terraform
data "aiven_cmk_accessor_azure" "example" {
  project = "my-project"

  /* COMPUTED FIELDS
  app_id = "foo"
  */
}
```

## Schema

### Required

- `project` (String) Project name.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `app_id` (String) The Azure application ID that Aiven uses to access your Key Vault.
- `id` (String) Resource ID, equal to `project`.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
