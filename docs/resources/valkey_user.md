---
page_title: "aiven_valkey_user Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an Aiven for Valkey™ https://aiven.io/docs/products/valkey service user. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.
---

# aiven_valkey_user (Resource)

Creates and manages an [Aiven for Valkey™](https://aiven.io/docs/products/valkey) service user. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.

## Example Usage

```terraform
resource "aiven_valkey_user" "example" {
  project      = "my-project" // Force new
  service_name = "my-valkey" // Force new
  username     = "testuser" // Force new

  // OPTIONAL FIELDS
  password_wo           = "password123"
  password_wo_version   = 1
  valkey_acl_categories = ["+@write", "+@keyspace"]
  valkey_acl_channels   = ["some*chan"]
  valkey_acl_commands   = ["+set", "+del", "+expire", "-flushall", "-flushdb"]
  valkey_acl_keys       = ["session:*"]

  /* COMPUTED FIELDS
  password_encryption_type = "md5"
  type                     = "foo"
  */
}
```

## Schema

### Required

- `project` (String) Project name. Changing this property forces recreation of the resource.
- `service_name` (String) Service name. Changing this property forces recreation of the resource.
- `username` (String) Service username. Maximum length: `64`. Must match pattern: `^[_A-Za-z0-9][-._A-Za-z0-9]{0,63}$`. Changing this property forces recreation of the resource.

### Optional

> **NOTE**: [Write-only arguments](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments) are supported in Terraform 1.11 and later.

- `password` (String, Sensitive) The password of the service user (auto-generated if not provided). The field conflicts with `password_wo`. Length must be between `8` and `256`.
- `password_wo` (String, Sensitive, [Write-only](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments)) The password of the service user (write-only, not stored in state). The field is required with `password_wo_version`. The field conflicts with `password`. Length must be between `8` and `256`.
- `password_wo_version` (Number) Version number for `password_wo`. Increment this to rotate the password. The field is required with `password_wo`. Minimum value: `1`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `valkey_acl_categories` (List of String) Allow or disallow command categories. To allow a category use the prefix `+@` and to disallow use `-@`. See the [Valkey documentation](https://valkey.io/topics/acl/) for details on the ACL feature. The field is required with `valkey_acl_commands` and `valkey_acl_keys`.
- `valkey_acl_channels` (List of String) Allows and disallows access to pub/sub channels. Entries are defined as standard glob patterns.
- `valkey_acl_commands` (List of String) Defines rules for individual commands. To allow a command use the prefix `+` and to disallow use `-`. The field is required with `valkey_acl_categories` and `valkey_acl_keys`.
- `valkey_acl_keys` (List of String) Key access rules. Entries are defined as standard glob patterns. The field is required with `valkey_acl_categories` and `valkey_acl_commands`.

### Read-Only

- `id` (String) Resource ID composed as: `project/service_name/username`.
- `password_encryption_type` (String) The password hashing algorithm used for this PostgreSQL user, derived from the stored password hash. 'unknown' is reported when the hash is missing or uses an unrecognised format. The possible values are `md5`, `scram-sha-256` and `unknown`.
- `type` (String) Account type.

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
terraform import aiven_valkey_user.example PROJECT/SERVICE_NAME/USERNAME
```
