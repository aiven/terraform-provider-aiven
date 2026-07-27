---
page_title: "aiven_valkey_user Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about an Aiven for Valkey™ service user.
---

# aiven_valkey_user (Data Source)

Gets information about an Aiven for Valkey™ service user.

## Example Usage

```terraform
data "aiven_valkey_user" "example" {
  project      = "my-project"
  service_name = "my-valkey"
  username     = "testuser"

  /* COMPUTED FIELDS
  password              = "password123"
  type                  = "foo"
  valkey_acl_categories = ["+@write", "+@keyspace"]
  valkey_acl_channels   = ["some*chan"]
  valkey_acl_commands   = ["+set", "+del", "+expire", "-flushall", "-flushdb"]
  valkey_acl_keys       = ["session:*"]
  */
}
```

## Schema

### Required

- `project` (String) Project name.
- `service_name` (String) Service name.
- `username` (String) Service username.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) Resource ID composed as: `project/service_name/username`.
- `password` (String, Sensitive) The password of the service user (auto-generated if not provided). The field conflicts with `password_wo`.
- `type` (String) Account type.
- `valkey_acl_categories` (List of String) Allow or disallow command categories. To allow a category use the prefix `+@` and to disallow use `-@`. See the [Valkey documentation](https://valkey.io/topics/acl/) for details on the ACL feature. The field is required with `valkey_acl_commands` and `valkey_acl_keys`.
- `valkey_acl_channels` (List of String) Allows and disallows access to pub/sub channels. Entries are defined as standard glob patterns.
- `valkey_acl_commands` (List of String) Defines rules for individual commands. To allow a command use the prefix `+` and to disallow use `-`. The field is required with `valkey_acl_categories` and `valkey_acl_keys`.
- `valkey_acl_keys` (List of String) Key access rules. Entries are defined as standard glob patterns. The field is required with `valkey_acl_categories` and `valkey_acl_commands`.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
