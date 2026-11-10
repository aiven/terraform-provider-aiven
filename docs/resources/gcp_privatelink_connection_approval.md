---
page_title: "aiven_gcp_privatelink_connection_approval Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Approve a new Google Private Service Connect connection. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.
---

# aiven_gcp_privatelink_connection_approval (Resource)

Approve a new Google Private Service Connect connection. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.

## Example Usage

```terraform
resource "aiven_gcp_privatelink_connection_approval" "example" {
  project         = aiven_gcp_privatelink.example.project      // e.g. "my-project". Force new
  service_name    = aiven_gcp_privatelink.example.service_name // e.g. "foo". Force new
  user_ip_address = "10.20.30.40"

  // OPTIONAL FIELDS
  psc_connection_id = "foo"

  /* COMPUTED FIELDS
  privatelink_connection_id = "foo"
  state                     = "active"
  */
}

resource "aiven_gcp_privatelink" "example" {
  project      = "my-project" // Force new
  service_name = "foo"        // Force new
}
```

## Schema

### Required

- `project` (String) Project name. Changing this property forces recreation of the resource.
- `service_name` (String) Service name. Changing this property forces recreation of the resource.
- `user_ip_address` (String) Frontend IP address of the forwarding rule that connects to the service at customer end. Must be a valid IPv4 address.

### Optional

- `psc_connection_id` (String) Google Private Service Connect connection ID for this connection. Changing to a different nonempty ID forces recreation of the resource. An empty string keeps the currently managed connection.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) Resource ID composed as: `project/service_name`.
- `privatelink_connection_id` (String) Privatelink connection ID.
- `state` (String) The Aiven connection state. The Google endpoint status has a separate value. The possible values are `active`, `connected`, `pending-user-approval` and `user-approved`.

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
terraform import aiven_gcp_privatelink_connection_approval.example PROJECT/SERVICE_NAME
```

With Terraform 1.12 or later, you can also import using resource identity:

```terraform
import {
  to = aiven_gcp_privatelink_connection_approval.example
  identity = {
    project      = "PROJECT"
    service_name = "SERVICE_NAME"
  }
}
```
