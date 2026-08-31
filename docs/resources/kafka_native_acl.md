---
page_title: "aiven_kafka_native_acl Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages Kafka-native access control lists https://aiven.io/docs/products/kafka/concepts/acl (ACLs) for an Aiven for Apache Kafka® service. ACLs control access to Kafka topics, consumer groups, clusters, and Schema Registry. Kafka-native ACLs provide advanced resource-level access control with fine-grained permissions, including ALLOW and DENY rules. For simplified topic-level control you can use Aiven ACLs https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/kafka_acl. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.
---

# aiven_kafka_native_acl (Resource)

Creates and manages Kafka-native [access control lists](https://aiven.io/docs/products/kafka/concepts/acl) (ACLs) for an Aiven for Apache Kafka® service. ACLs control access to Kafka topics, consumer groups, clusters, and Schema Registry.

Kafka-native ACLs provide advanced resource-level access control with fine-grained permissions, including `ALLOW` and `DENY` rules. For simplified topic-level control you can use [Aiven ACLs](https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/kafka_acl). If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.

## Example Usage

```terraform
resource "aiven_kafka_native_acl" "example" {
  project         = "my-project" // Force new
  service_name    = "my-kafka" // Force new
  operation       = "Read" // Force new
  pattern_type    = "LITERAL" // Force new
  permission_type = "ALLOW" // Force new
  principal       = "User:alice" // Force new
  resource_name   = "consumer-group-1." // Force new
  resource_type   = "Topic" // Force new

  // OPTIONAL FIELDS
  host = "*" // Force new

  /* COMPUTED FIELDS
  acl_id = "foo"
  */
}
```

## Schema

### Required

- `operation` (String) Kafka ACL operation represents an operation which an ACL grants or denies permission to perform. The possible values are `All`, `Alter`, `AlterConfigs`, `ClusterAction`, `Create`, `CreateTokens`, `Delete`, `Describe`, `DescribeConfigs`, `DescribeTokens`, `IdempotentWrite`, `Read` and `Write`. Changing this property forces recreation of the resource.
- `pattern_type` (String) How a Kafka-native ACL matches its resource name. The possible values are `LITERAL` and `PREFIXED`. Changing this property forces recreation of the resource.
- `permission_type` (String) Whether a Kafka-native ACL allows or denies its operation. The possible values are `ALLOW` and `DENY`. Changing this property forces recreation of the resource.
- `principal` (String) principal is in 'principalType:name' format. Maximum length: `256`. Changing this property forces recreation of the resource.
- `project` (String) Project name. Changing this property forces recreation of the resource.
- `resource_name` (String) Resource pattern used to match specified resources. Maximum length: `256`. Changing this property forces recreation of the resource.
- `resource_type` (String) Kafka ACL resource type represents a type of resource which an ACL can be applied to. The possible values are `Cluster`, `DelegationToken`, `Group`, `Topic`, `TransactionalId` and `User`. Changing this property forces recreation of the resource.
- `service_name` (String) Service name. Changing this property forces recreation of the resource.

### Optional

- `host` (String) the host or `*` for all hosts. Maximum length: `256`. The default value is `*`. Changing this property forces recreation of the resource.
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
terraform import aiven_kafka_native_acl.example PROJECT/SERVICE_NAME/ACL_ID
```
