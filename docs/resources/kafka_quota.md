---
page_title: "aiven_kafka_quota Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages quotas for an Aiven for Apache Kafka® service user. If a refresh can't find this resource, it's removed from the state and a new create plan is generated.
---

# aiven_kafka_quota (Resource)

Creates and manages quotas for an Aiven for Apache Kafka® service user. If a refresh can't find this resource, it's removed from the state and a new create plan is generated.

## Example Usage

```terraform
resource "aiven_kafka_quota" "example" {
  project      = aiven_kafka.example_kafka.project      // e.g. "my-project". Force new
  service_name = aiven_kafka.example_kafka.service_name // e.g. "my-kafka". Force new

  // OPTIONAL FIELDS
  client_id          = "abc123" // Force new
  user               = "alice"  // Force new
  consumer_byte_rate = 1073741824
  producer_byte_rate = 1073741824
  request_percentage = 100
}

resource "aiven_kafka" "example_kafka" {
  project      = "my-project"
  plan         = "business-4"
  service_name = "example-kafka"
}
```

## Schema

### Required

- `project` (String) Project name. Changing this property forces recreation of the resource.
- `service_name` (String) Service name. Changing this property forces recreation of the resource.

### Optional

- `client_id` (String) Represents a logical group of clients, assigned a unique name by the client application. Quotas can be applied based on user, client-id, or both. The most relevant quota is chosen for each connection. All connections within a quota group share the same quota. At least one of the fields must be specified: `client_id` or `user`. Maximum length: `255`. Changing this property forces recreation of the resource.
- `consumer_byte_rate` (Number) Defines the bandwidth limit in bytes/sec for each group of clients sharing a quota. Every distinct client group is allocated a specific quota, as defined by the cluster, on a per-broker basis. Exceeding this limit results in client throttling. At least one of the fields must be specified: `consumer_byte_rate`, `producer_byte_rate` or `request_percentage`. Maximum value: `1073741824`.
- `producer_byte_rate` (Number) Defines the bandwidth limit in bytes/sec for each group of clients sharing a quota. Every distinct client group is allocated a specific quota, as defined by the cluster, on a per-broker basis. Exceeding this limit results in client throttling. At least one of the fields must be specified: `consumer_byte_rate`, `producer_byte_rate` or `request_percentage`. Maximum value: `1073741824`.
- `request_percentage` (Number) Sets the maximum percentage of CPU time that a client group can use on request handler I/O and network threads per broker within a quota window. Exceeding this limit triggers throttling. The quota, expressed as a percentage, also indicates the total allowable CPU usage for the client groups sharing the quota. At least one of the fields must be specified: `consumer_byte_rate`, `producer_byte_rate` or `request_percentage`. Maximum value: `100`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `user` (String) Represents a logical group of clients, assigned a unique name by the client application. Quotas can be applied based on user, client-id, or both. The most relevant quota is chosen for each connection. All connections within a quota group share the same quota. At least one of the fields must be specified: `client_id` or `user`. Maximum length: `64`. Must match pattern: `^(\*$|[a-zA-Z0-9_?][a-zA-Z0-9-_?*.].{0,62})$`. Changing this property forces recreation of the resource.

### Read-Only

- `id` (String) Resource ID composed as: `project/service_name/client_id/user`.

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
terraform import aiven_kafka_quota.example PROJECT/SERVICE_NAME/CLIENT_ID/USER
```

With Terraform 1.12 or later, you can also import using resource identity:

```terraform
import {
  to = aiven_kafka_quota.example
  identity = {
    project      = "PROJECT"
    service_name = "SERVICE_NAME"
    client_id    = "CLIENT_ID" # Optional
    user         = "USER"      # Optional
  }
}
```
