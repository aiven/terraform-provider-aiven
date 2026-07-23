---
page_title: "aiven_organization_user_group Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about an existing user group in an organization.
---

# aiven_organization_user_group (Data Source)

Gets information about an existing user group in an organization.

## Example Usage

```terraform
data "aiven_organization_user_group" "example" {
  organization_id = "org1a23f456789"

  // LOOKUP — provide `group_id` or `name`
  group_id = "foo"
  // name  = "Admin Users"

  /* COMPUTED FIELDS
  create_time     = "2021-01-01T00:00:00Z"
  description     = "The group of admins for the organization"
  managed_by_scim = true
  update_time     = "2021-01-01T00:00:00Z"
  */
}
```

## Schema

### Required

- `organization_id` (String) ID of an organization.

### Optional

- `group_id` (String) ID of the user group. Exactly one of the fields must be specified: `group_id` or `name`.
- `name` (String) User Group Name. Exactly one of the fields must be specified: `group_id` or `name`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `create_time` (String) User group creation time.
- `description` (String) Description.
- `id` (String) Resource ID composed as: `organization_id/group_id`.
- `managed_by_scim` (Boolean) Managed By Scim.
- `update_time` (String) User group last update time.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
