---
page_title: "aiven_transit_gateway_vpc_attachment Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an AWS Transit Gateway VPC attachment for an Aiven project VPC. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.
---

# aiven_transit_gateway_vpc_attachment (Resource)

Creates and manages an AWS Transit Gateway VPC attachment for an Aiven project VPC. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.

## Example Usage

```terraform
resource "aiven_transit_gateway_vpc_attachment" "example" {
  peer_cloud_account = "123456789012"          // Force new
  peer_vpc           = "tgw-0123456789abcdef0" // Force new

  // OPTIONAL FIELDS
  project                 = aiven_project_vpc.example.project        // e.g. "example-project"
  project_vpc_id          = aiven_project_vpc.example.project_vpc_id // e.g. "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  peer_region             = "us-east-1"                              // Force new
  user_peer_network_cidrs = ["192.168.6.0/24"]

  /* COMPUTED FIELDS
  peering_connection_id = "pcx-0123456789abcdef0"
  state                 = "ACTIVE"
  state_info = {
    foo = "foo"
  }
  */
}

resource "aiven_project_vpc" "example" {
  project      = "my-project"       // Force new
  cloud_name   = "aws-eu-central-1" // Force new
  network_cidr = "192.168.6.0/24"   // Force new
}
```

## Schema

### Required

- `peer_cloud_account` (String) AWS account ID that owns the Transit Gateway. Maximum length: `1024`. Changing this property forces recreation of the resource.
- `peer_vpc` (String) AWS Transit Gateway ID. Maximum length: `1024`. Changing this property forces recreation of the resource.

### Optional

- `peer_region` (String) AWS region of the Transit Gateway. When omitted, the Aiven project VPC region is used. Maximum length: `1024`. Changing this property forces recreation of the resource.
- `project` (String) The name of the project the Aiven VPC belongs to. Changing the Aiven VPC forces recreation of the resource. The field is required with `project_vpc_id`.
- `project_vpc_id` (String) The ID of the Aiven VPC. Changing the Aiven VPC forces recreation of the resource. The field is required with `project`. The field conflicts with `vpc_id`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `user_peer_network_cidrs` (Set of String) List of private IPv4 ranges to route through the peering connection.
- `vpc_id` (String, Deprecated) Aiven project VPC ID in the `PROJECT/PROJECT_VPC_ID` format. Changing the Aiven VPC forces recreation of the resource; switching to `project` and `project_vpc_id` doesn't. Exactly one of the fields must be specified: `project` or `vpc_id`. Must match pattern: `^[^/]*/[^/]*$`. **Deprecated**: Use the separate `project` and `project_vpc_id` fields instead.

### Read-Only

- `id` (String) Resource ID composed as: `project/project_vpc_id/peer_cloud_account/peer_vpc`.
- `peering_connection_id` (String) Legacy AWS VPC peering connection ID (`pcx-*`) for ordinary AWS VPC peering connections, if available. This is not the AWS Transit Gateway attachment ID; TGW attachment details are exposed in `state_info`.
- `state` (String) Project VPC peering connection state. The possible values are `ACTIVE`, `APPROVED`, `APPROVED_PEER_REQUESTED`, `DELETED`, `DELETED_BY_PEER`, `DELETING`, `ERROR`, `INVALID_SPECIFICATION`, `PENDING_PEER` and `REJECTED_BY_PEER`.
- `state_info` (Map of String) State-specific help or error information.

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
terraform import aiven_transit_gateway_vpc_attachment.example PROJECT/PROJECT_VPC_ID/PEER_CLOUD_ACCOUNT/PEER_VPC
```
