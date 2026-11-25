---
page_title: "aiven_kafka_schema_configuration Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  The Kafka Schema Configuration resource allows the creation and management of Aiven Kafka Schema Configurations. If a refresh can't find this resource, it's removed from the state and a new create plan is generated.
---

# aiven_kafka_schema_configuration (Resource)

The Kafka Schema Configuration resource allows the creation and management of Aiven Kafka Schema Configurations. If a refresh can't find this resource, it's removed from the state and a new create plan is generated.

## Example Usage

```terraform
resource "aiven_kafka_schema_configuration" "example" {
  project             = aiven_kafka.example_kafka.project
  service_name        = aiven_kafka.example_kafka.service_name
  compatibility_level = "FULL"
}

resource "aiven_kafka" "example_kafka" {
  project      = "my-project"
  plan         = "business-4"
  service_name = "example-kafka"

  kafka_user_config {
    schema_registry = true
  }
}
```

## Schema

### Required

- `project` (String) Project name. Changing this property forces recreation of the resource.
- `service_name` (String) Service name. Changing this property forces recreation of the resource.

### Optional

- `compatibility_level` (String) Compatibility level. The possible values are `BACKWARD`, `BACKWARD_TRANSITIVE`, `FORWARD`, `FORWARD_TRANSITIVE`, `FULL`, `FULL_TRANSITIVE` and `NONE`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) Resource ID composed as: `project/service_name`.

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
terraform import aiven_kafka_schema_configuration.example PROJECT/SERVICE_NAME
```

With Terraform 1.12 or later, you can also import using resource identity:

```terraform
import {
  to = aiven_kafka_schema_configuration.example
  identity = {
    project      = "PROJECT"
    service_name = "SERVICE_NAME"
  }
}
```
