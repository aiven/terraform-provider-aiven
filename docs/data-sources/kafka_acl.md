---
page_title: "aiven_kafka_acl Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about an ACL entry for an Aiven for Apache Kafka® service.
---

# aiven_kafka_acl (Data Source)

Gets information about an ACL entry for an Aiven for Apache Kafka® service.

## Example Usage

```terraform
data "aiven_kafka_acl" "example" {
  project      = "my-project"
  service_name = "my-kafka"

  // LOOKUP — provide `acl_id`, or all of: `permission`, `topic` and `username`
  acl_id        = "foo"
  // permission = "readwrite"
  // topic      = "top*"
  // username   = "admin*"
}
```

## Schema

### Required

- `project` (String) Project name.
- `service_name` (String) Service name.

### Optional

- `acl_id` (String) Kafka ACL ID. Provide either `acl_id`, or all of `permission`, `topic` and `username` together.
- `permission` (String) Permission of an Aiven Kafka ACL entry, as opposed to a Kafka-native one. The possible values are `admin`, `read`, `readwrite` and `write`. Provide either `acl_id`, or all of `permission`, `topic` and `username` together.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `topic` (String) Topic name pattern. Provide either `acl_id`, or all of `permission`, `topic` and `username` together.
- `username` (String) Username. Provide either `acl_id`, or all of `permission`, `topic` and `username` together.

### Read-Only

- `id` (String) Resource ID composed as: `project/service_name/acl_id`.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
