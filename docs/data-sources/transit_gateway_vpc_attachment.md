---
page_title: "aiven_transit_gateway_vpc_attachment Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about an AWS Transit Gateway VPC attachment for an Aiven project VPC.
---

# aiven_transit_gateway_vpc_attachment (Data Source)

Gets information about an AWS Transit Gateway VPC attachment for an Aiven project VPC.

## Example Usage

```terraform
data "aiven_transit_gateway_vpc_attachment" "example" {
  vpc_id             = "example-project/example-vpc"
  peer_cloud_account = "123456789012"
  peer_vpc           = "tgw-0123456789abcdef0"

  // OPTIONAL FIELDS
  peer_region = "us-east-1"

  /* COMPUTED FIELDS
  id                    = "example-project/example-vpc/123456789012/tgw-0123456789abcdef0/us-east-1"
  peering_connection_id = "pcx-0123456789abcdef0"
  state                 = "ACTIVE"
  state_info = {
    foo = "foo"
  }
  user_peer_network_cidrs = ["192.168.6.0/24"]
  */
}
```

## Schema

### Required

- `peer_cloud_account` (String) AWS account ID that owns the Transit Gateway.
- `peer_vpc` (String) AWS Transit Gateway ID.
- `vpc_id` (String) Aiven project VPC ID in the `PROJECT/VPC_ID` format.

### Optional

- `peer_region` (String) AWS region of the Transit Gateway. When omitted, the data source searches all regions and requires a single matching attachment.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) Terraform identifier for the VPC peering connection.
- `peering_connection_id` (String) Legacy AWS VPC peering connection ID (`pcx-*`) for ordinary AWS VPC peering connections, if available. This is not the AWS Transit Gateway attachment ID; TGW attachment details are exposed in `state_info`.
- `state` (String) Project VPC peering connection state. The possible values are `ACTIVE`, `APPROVED`, `APPROVED_PEER_REQUESTED`, `DELETED`, `DELETED_BY_PEER`, `DELETING`, `ERROR`, `INVALID_SPECIFICATION`, `PENDING_PEER` and `REJECTED_BY_PEER`.
- `state_info` (Map of String) State-specific help or error information.
- `user_peer_network_cidrs` (Set of String) List of private IPv4 ranges to route through the peering connection.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
