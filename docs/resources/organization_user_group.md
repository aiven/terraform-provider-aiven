---
page_title: "aiven_organization_user_group Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages a user group https://aiven.io/docs/platform/howto/list-groups in an organization.
---

# aiven_organization_user_group (Resource)

Creates and manages a [user group](https://aiven.io/docs/platform/howto/list-groups) in an organization.

## Example Usage

```terraform
resource "aiven_organization_user_group" "example" {
  organization_id = "org1a23f456789" // Force new
  description     = "The group of admins for the organization"
  name            = "Admin Users"

  /* COMPUTED FIELDS
  group_id        = "foo"
  create_time     = "2021-01-01T00:00:00Z"
  managed_by_scim = true
  update_time     = "2021-01-01T00:00:00Z"
  */
}
```

## Schema

### Required

- `description` (String) Description. Maximum length: `4096`.
- `name` (String) User Group Name. Maximum length: `128`.
- `organization_id` (String) ID of an organization. Changing this property forces recreation of the resource.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `create_time` (String) User group creation time.
- `group_id` (String) ID of the user group.
- `id` (String) Resource ID composed as: `organization_id/group_id`.
- `managed_by_scim` (Boolean) Managed By Scim.
- `update_time` (String) User group last update time.

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
terraform import aiven_organization_user_group.example ORGANIZATION_ID/GROUP_ID
```
