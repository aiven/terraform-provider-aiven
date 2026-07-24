---
page_title: "aiven_gcp_privatelink Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about a Google Private Service Connect connection for an Aiven service.
---

# aiven_gcp_privatelink (Data Source)

Gets information about a Google Private Service Connect connection for an Aiven service.

## Example Usage

```terraform
data "aiven_gcp_privatelink" "example" {
  project      = "my-project"
  service_name = "foo"

  /* COMPUTED FIELDS
  google_service_attachment = "foo"
  state                     = "active"
  */
}
```

## Schema

### Required

- `project` (String) Project name.
- `service_name` (String) Service name.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `google_service_attachment` (String) Google Private Service Connect service attachment.
- `id` (String) Resource ID composed as: `project/service_name`.
- `message` (String, Deprecated) Legacy response message retained for backward compatibility. **Deprecated**: This attribute is retained only for compatibility with state created by older provider versions and is no longer populated.
- `state` (String) The state of the Private Service Connect resource. The possible values are `active`, `creating` and `deleting`.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
