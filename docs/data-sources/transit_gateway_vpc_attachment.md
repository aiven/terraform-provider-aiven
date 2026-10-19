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
  peer_cloud_account = "123456789012"
  peer_vpc           = "tgw-0123456789abcdef0"

  // OPTIONAL FIELDS
  project        = "example-project"
  project_vpc_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  peer_region    = "us-east-1"

  /* COMPUTED FIELDS
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

### Optional

- `peer_region` (String) AWS region of the Transit Gateway. When omitted, the data source searches all regions and requires a single matching attachment.
- `project` (String) The name of the project the Aiven VPC belongs to. The field is required with `project_vpc_id`.
- `project_vpc_id` (String) The ID of the Aiven VPC. The field is required with `project`. The field conflicts with `vpc_id`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `vpc_id` (String, Deprecated) Aiven project VPC ID in the `PROJECT/PROJECT_VPC_ID` format. Exactly one of the fields must be specified: `project` or `vpc_id`. **Deprecated**: Use the separate `project` and `project_vpc_id` fields instead.

### Read-Only

- `id` (String) Resource ID composed as: `project/project_vpc_id/peer_cloud_account/peer_vpc`.
- `peering_connection_id` (String) Legacy AWS VPC peering connection ID (`pcx-*`) for ordinary AWS VPC peering connections, if available. This is not the AWS Transit Gateway attachment ID; TGW attachment details are exposed in `state_info`.
- `state` (String) Project VPC peering connection state. The possible values are `ACTIVE`, `APPROVED`, `APPROVED_PEER_REQUESTED`, `DELETED`, `DELETED_BY_PEER`, `DELETING`, `ERROR`, `INVALID_SPECIFICATION`, `PENDING_PEER` and `REJECTED_BY_PEER`.
- `state_info` (Map of String) State-specific help or error information.
- `user_peer_network_cidrs` (Set of String) List of private IPv4 ranges to route through the peering connection.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
