# Transit Gateway VPC Attachment Data Source

The Transit Gateway VPC Attachment resource allows the creation and management Transit 
Gateway VPC Attachment VPC peering connection between Aiven and AWS.  

## Example Usage

```hcl
data "aiven_transit_gateway_vpc_attachment" "attachment" {
    vpc_id = aiven_project_vpc.bar.id
    peer_cloud_account = "<PEER_ACCOUNT_ID>"
    peer_vpc = "google-project1"
}
```

## Argument Reference

* `vpc_id` - (Required) is the Aiven VPC the peering connection is associated with.

* `peer_cloud_account` - (Required) AWS account ID of the peered VPC.

* `peer_vpc` - (Required) Transit gateway ID

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `peer_region` - AWS region of the peered VPC (if not in the same region as Aiven VPC).

* `user_peer_network_cidrs` - List of private IPv4 ranges to route through the peering connection.

* `state` is the state of the peering connection.

* `state_info` state-specific help or error information.

* `peering_connection_id` Cloud provider identifier for the peering connection if available.
