---
page_title: "aiven_opensearch_user Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about an Aiven for OpenSearch® service user.
---

# aiven_opensearch_user (Data Source)

Gets information about an Aiven for OpenSearch® service user.

## Example Usage

```terraform
data "aiven_opensearch_user" "example" {
  project      = "my-project"
  service_name = "my-opensearch"
  username     = "testuser"

  /* COMPUTED FIELDS
  password                 = "password123"
  password_encryption_type = "md5"
  type                     = "foo"
  */
}
```

## Schema

### Required

- `project` (String) Project name.
- `service_name` (String) Service name.
- `username` (String) Account username.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) Resource ID composed as: `project/service_name/username`.
- `password` (String, Sensitive) The password of the service user (auto-generated if not provided). The field conflicts with `password_wo`.
- `password_encryption_type` (String) The password hashing algorithm used for this PostgreSQL user, derived from the stored password hash. 'unknown' is reported when the hash is missing or uses an unrecognised format. The possible values are `md5`, `scram-sha-256` and `unknown`.
- `type` (String) Account type.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
