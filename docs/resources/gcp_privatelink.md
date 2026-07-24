---
page_title: "aiven_gcp_privatelink Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages a Google Private Service Connect for an Aiven service in a VPC. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.
---

# aiven_gcp_privatelink (Resource)

Creates and manages a Google Private Service Connect for an Aiven service in a VPC. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.

## Example Usage

```terraform
resource "aiven_gcp_privatelink" "example" {
  project      = "my-project" // Force new
  service_name = "foo" // Force new

  /* COMPUTED FIELDS
  google_service_attachment = "foo"
  state                     = "active"
  */
}
```

## Schema

### Required

- `project` (String) Project name. Changing this property forces recreation of the resource.
- `service_name` (String) Service name. Changing this property forces recreation of the resource.

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

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `default` (String, Deprecated) Timeout for all operations. Deprecated, use operation-specific timeouts instead.
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

## Import

Import is supported using the following syntax:

```shell
terraform import aiven_gcp_privatelink.example PROJECT/SERVICE_NAME
```
