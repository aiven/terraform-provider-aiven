---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_transit_gateway_vpc_attachment Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  The Transit Gateway VPC Attachment resource allows the creation and management Transit Gateway VPC Attachment VPC peering connection between Aiven and AWS.
---

# aiven_transit_gateway_vpc_attachment (Resource)

The Transit Gateway VPC Attachment resource allows the creation and management Transit Gateway VPC Attachment VPC peering connection between Aiven and AWS.

## Example Usage

```terraform
resource "aiven_transit_gateway_vpc_attachment" "attachment" {
  vpc_id             = aiven_project_vpc.bar.id
  peer_cloud_account = "<PEER_ACCOUNT_ID>"
  peer_vpc           = "google-project1"
  peer_region        = "aws-eu-west-1"
  user_peer_network_cidrs = [
    "10.0.0.0/24"
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `peer_cloud_account` (String) AWS account ID or GCP project ID of the peered VPC. Changing this property forces recreation of the resource.
- `peer_vpc` (String) Transit gateway ID. Changing this property forces recreation of the resource.
- `user_peer_network_cidrs` (Set of String) List of private IPv4 ranges to route through the peering connection
- `vpc_id` (String) The VPC the peering connection belongs to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.

### Optional

- `peer_region` (String) AWS region of the peered VPC (if not in the same region as Aiven VPC). This value can't be changed.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of this resource.
- `peering_connection_id` (String) Cloud provider identifier for the peering connection if available
- `state` (String) State of the peering connection
- `state_info` (Map of String) State-specific help or error information

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `default` (String)
- `delete` (String)
- `read` (String)
- `update` (String)

## Import

Import is supported using the following syntax:

```shell
terraform import aiven_transit_gateway_vpc_attachment.attachment PROJECT/VPC_ID/PEER_CLOUD_ACCOUNT/PEER_VPC/PEER_REGION
```
