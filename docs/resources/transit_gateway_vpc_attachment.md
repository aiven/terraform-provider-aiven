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
  vpc_id             = "example-project/example-vpc" // Force new
  peer_cloud_account = "123456789012" // Force new
  peer_vpc           = "tgw-0123456789abcdef0" // Force new

  // OPTIONAL FIELDS
  peer_region             = "us-east-1" // Force new
  user_peer_network_cidrs = ["192.168.6.0/24"]

  /* COMPUTED FIELDS
  id                    = "example-project/example-vpc/123456789012/tgw-0123456789abcdef0/us-east-1"
  peering_connection_id = "pcx-0123456789abcdef0"
  state                 = "ACTIVE"
  state_info = {
    foo = "foo"
  }
  */
}
```

## Schema

### Required

- `peer_cloud_account` (String) AWS account ID that owns the Transit Gateway. Maximum length: `1024`. Changing this property forces recreation of the resource.
- `peer_vpc` (String) AWS Transit Gateway ID. Maximum length: `1024`. Changing this property forces recreation of the resource.
- `vpc_id` (String) Aiven project VPC ID in the `PROJECT/VPC_ID` format. Changing this property forces recreation of the resource.

### Optional

- `peer_region` (String) AWS region of the Transit Gateway. When omitted, the Aiven project VPC region is used. Maximum length: `1024`. Changing this property forces recreation of the resource.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `user_peer_network_cidrs` (Set of String) List of private IPv4 ranges to route through the peering connection.

### Read-Only

- `id` (String) Terraform identifier for the VPC peering connection.
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

Import is supported using one of the following formats:

```shell
terraform import aiven_transit_gateway_vpc_attachment.example PROJECT/VPC_ID/PEER_CLOUD_ACCOUNT/PEER_VPC
terraform import aiven_transit_gateway_vpc_attachment.example PROJECT/VPC_ID/PEER_CLOUD_ACCOUNT/PEER_VPC/PEER_REGION
```
