---
page_title: "aiven_kafka_schema_configuration Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  The Kafka Schema Configuration data source provides information about the existing Aiven Kafka Schema Configuration.
---

# aiven_kafka_schema_configuration (Data Source)

The Kafka Schema Configuration data source provides information about the existing Aiven Kafka Schema Configuration.

## Example Usage

```terraform
data "aiven_kafka_schema_configuration" "example" {
  project      = "my-project"
  service_name = "my-kafka"

  /* COMPUTED FIELDS
  compatibility_level = "FULL"
  */
}
```

## Schema

### Required

- `project` (String) Project name.
- `service_name` (String) Service name.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `compatibility_level` (String) Compatibility level. The possible values are `BACKWARD`, `BACKWARD_TRANSITIVE`, `FORWARD`, `FORWARD_TRANSITIVE`, `FULL`, `FULL_TRANSITIVE` and `NONE`.
- `id` (String) Resource ID composed as: `project/service_name`.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
