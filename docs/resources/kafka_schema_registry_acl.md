---
page_title: "aiven_kafka_schema_registry_acl Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an Aiven for Apache Kafka® Schema Registry ACL entry. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.
---

# aiven_kafka_schema_registry_acl (Resource)

Creates and manages an Aiven for Apache Kafka® Schema Registry ACL entry. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.

## Example Usage

```terraform
resource "aiven_kafka_schema_registry_acl" "example" {
  project      = "my-project" // Force new
  service_name = "my-kafka" // Force new
  permission   = "schema_registry_read" // Force new
  resource     = "Config:" // Force new
  username     = "admin*" // Force new

  /* COMPUTED FIELDS
  acl_id = "foo"
  */
}
```

## Schema

### Required

- `permission` (String) ACL entry for Schema Registry. The possible values are `schema_registry_read` and `schema_registry_write`. Changing this property forces recreation of the resource.
- `project` (String) Project name. Changing this property forces recreation of the resource.
- `resource` (String) Schema Registry ACL entry resource name pattern. Length must be between `1` and `249`. Changing this property forces recreation of the resource.
- `service_name` (String) Service name. Changing this property forces recreation of the resource.
- `username` (String) Username. Length must be between `1` and `64`. Changing this property forces recreation of the resource.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `acl_id` (String) Kafka Schema Registry ACL ID.
- `id` (String) Resource ID composed as: `project/service_name/acl_id`.

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
terraform import aiven_kafka_schema_registry_acl.example PROJECT/SERVICE_NAME/ACL_ID
```
