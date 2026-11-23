---
page_title: "aiven_kafka_schema Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  The Kafka Schema resource allows the creation and management of Aiven Kafka Schemas. Schemas with references are hard deleted, while schemas without references are only soft deleted. If a refresh can't find this resource, it's removed from the state and a new create plan is generated.
---

# aiven_kafka_schema (Resource)

The Kafka Schema resource allows the creation and management of Aiven Kafka Schemas. Schemas with references are hard deleted, while schemas without references are only soft deleted. If a refresh can't find this resource, it's removed from the state and a new create plan is generated.

## Example Usage

```terraform
resource "aiven_kafka_schema" "example" {
  project             = aiven_kafka.example_kafka.project
  service_name        = aiven_kafka.example_kafka.service_name
  subject_name        = "my-topic-value"
  compatibility_level = "FULL"
  schema = jsonencode({
    fields = [{
      name = "id"
      type = "string"
    }]
    name = "Order"
    type = "record"
  })
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
- `schema` (String) Kafka Schema configuration. Should be a valid Avro, JSON, or Protobuf schema, depending on the schema type. Maximum length: `1048576`.
- `service_name` (String) Service name. Changing this property forces recreation of the resource.
- `subject_name` (String) Subject name. Changing this property forces recreation of the resource.

### Optional

- `compatibility_level` (String) Compatibility level. The possible values are `BACKWARD`, `BACKWARD_TRANSITIVE`, `FORWARD`, `FORWARD_TRANSITIVE`, `FULL`, `FULL_TRANSITIVE` and `NONE`.
- `references` (Block Set, Max: 128) Schema references (see [below for nested schema](#nestedblock--references))
- `schema_type` (String) Schema type, `AVRO` by default. Changing it forces recreation of the resource; removing it keeps the current type. The possible values are `AVRO`, `JSON` and `PROTOBUF`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) Resource ID composed as: `project/service_name/subject_name`.
- `version` (Number) Version.

<a id="nestedblock--references"></a>
### Nested Schema for `references`

Required:

- `name` (String) The name used to reference the provided subject and version. Maximum length: `1024`.
- `subject` (String) Subject. Maximum length: `1024`.
- `version` (Number) Version.


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
terraform import aiven_kafka_schema.example PROJECT/SERVICE_NAME/SUBJECT_NAME
```

With Terraform 1.12 or later, you can also import using resource identity:

```terraform
import {
  to = aiven_kafka_schema.example
  identity = {
    project      = "PROJECT"
    service_name = "SERVICE_NAME"
    subject_name = "SUBJECT_NAME"
  }
}
```
