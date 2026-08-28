---
page_title: "aiven_kafka_acl Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages Aiven access control lists https://aiven.io/docs/products/kafka/concepts/acl (ACLs) for an Aiven for Apache Kafka® service. ACLs control access to Kafka topics, consumer groups, clusters, and Schema Registry. Aiven ACLs provide simplified topic-level control with basic permissions and wildcard support. For more advanced access control, you can use Kafka-native ACLs https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/kafka_native_acl. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.
---

# aiven_kafka_acl (Resource)

Creates and manages Aiven [access control lists](https://aiven.io/docs/products/kafka/concepts/acl) (ACLs) for an Aiven for Apache Kafka® service. ACLs control access to Kafka topics, consumer groups, clusters, and Schema Registry.

Aiven ACLs provide simplified topic-level control with basic permissions and wildcard support. For more advanced access control, you can use [Kafka-native ACLs](https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/kafka_native_acl). If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.

## Example Usage

```terraform
resource "aiven_kafka_acl" "example" {
  project      = "my-project" // Force new
  service_name = "my-kafka" // Force new
  permission   = "readwrite" // Force new
  topic        = "top*" // Force new
  username     = "admin*" // Force new

  /* COMPUTED FIELDS
  acl_id = "foo"
  */
}
```

## Schema

### Required

- `permission` (String) Permission of an Aiven Kafka ACL entry, as opposed to a Kafka-native one. The possible values are `admin`, `read`, `readwrite` and `write`. Changing this property forces recreation of the resource.
- `project` (String) Project name. Changing this property forces recreation of the resource.
- `service_name` (String) Service name. Changing this property forces recreation of the resource.
- `topic` (String) Topic name pattern. Length must be between `1` and `249`. Changing this property forces recreation of the resource.
- `username` (String) Username. Length must be between `1` and `64`. Must match pattern: `^[-._*?A-Za-z0-9]+$`. Changing this property forces recreation of the resource.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `acl_id` (String) Kafka ACL ID.
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
terraform import aiven_kafka_acl.example PROJECT/SERVICE_NAME/ACL_ID
```
