---
page_title: "aiven_cmk_accessor_oci Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets the OCI CMK accessor for an Aiven project. The accessor is used to authenticate Aiven to your Oracle Cloud Infrastructure (OCI) Vault.
---

# aiven_cmk_accessor_oci (Data Source)

Gets the OCI CMK accessor for an Aiven project. The accessor is used to authenticate Aiven to your Oracle Cloud Infrastructure (OCI) Vault.

## Example Usage

```terraform
data "aiven_cmk_accessor_oci" "example" {
  project = "my-project"

  /* COMPUTED FIELDS
  access_group  = "foo"
  access_tenant = "foo"
  */
}
```

## Schema

### Required

- `project` (String) Project name.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `access_group` (String) The OCI access group that Aiven uses to access your OCI Vault key.
- `access_tenant` (String) The OCI access tenant that Aiven uses to access your OCI Vault key.
- `id` (String) Resource ID, equal to `project`.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
