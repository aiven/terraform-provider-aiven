---
page_title: "aiven_kafka_schema Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  The Kafka Schema data source provides information about the existing Aiven Kafka Schema.
---

# aiven_kafka_schema (Data Source)

The Kafka Schema data source provides information about the existing Aiven Kafka Schema.

## Example Usage

```terraform
data "aiven_kafka_schema" "example" {
  project      = "my-project"
  service_name = "my-kafka"
  subject_name = "my-topic-value"

  /* COMPUTED FIELDS
  compatibility_level = "FULL"
  references {
    name    = "container1.proto"
    subject = "topic1-values"
    version = 42
  }
  schema = jsonencode({
    fields = [{
      name = "id"
      type = "string"
    }]
    name = "Order"
    type = "record"
  })
  schema_type = "AVRO"
  version     = 42
  */
}
```

## Schema

### Required

- `project` (String) Project name.
- `service_name` (String) Service name.
- `subject_name` (String) Subject name.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `compatibility_level` (String) Compatibility level. The possible values are `BACKWARD`, `BACKWARD_TRANSITIVE`, `FORWARD`, `FORWARD_TRANSITIVE`, `FULL`, `FULL_TRANSITIVE` and `NONE`.
- `id` (String) Resource ID composed as: `project/service_name/subject_name`.
- `references` (Block Set) Schema references (see [below for nested schema](#nestedblock--references))
- `schema` (String) Kafka Schema configuration. Should be a valid Avro, JSON, or Protobuf schema, depending on the schema type.
- `schema_type` (String) Schema type, `AVRO` by default. Changing it forces recreation of the resource; removing it keeps the current type. The possible values are `AVRO`, `JSON` and `PROTOBUF`.
- `version` (Number) Version.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedblock--references"></a>
### Nested Schema for `references`

Read-Only:

- `name` (String) The name used to reference the provided subject and version.
- `subject` (String) Subject.
- `version` (Number) Version.
