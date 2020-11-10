# VPC Peering Connection Resource

The VPC Peering Connection resource allows the creation and management of Aiven VPC Peering Connections.

## Example Usage

```hcl
resource "aiven_vpc_peering_connection" "mypeeringconnection" {
    vpc_id = "${aiven_project_vpc.myvpc.id}"
    peer_cloud_account = "<PEER_ACCOUNT_ID>"
    peer_vpc = "<PEER_VPC_ID/NAME>"
    peer_region = "<PEER_REGION>"

    timeouts {
        create = "10m"
    }
}
```

## Argument Reference

* `vpc_id` - (Required) is the Aiven VPC the peering connection is associated with.

* `peer_cloud_account` - (Required) defines the identifier of the cloud account the VPC is being
peered with.

* `peer_vpc` - (Required) defines the identifier or name of the remote VPC.

* `peer_region` - (Optional) defines the region of the remote VPC if it is not in the same region as Aiven VPC.

* `timeouts` - (Optional) a custom client timeouts.

* `peering_connection_id` - (Optional) a cloud provider identifier for the peering connection if available.

* `peer_azure_app_id` - (Optional) an Azure app registration id in UUID4 form that is allowed to create a peering to the peer vnet. 

* `peer_azure_tenant_id` - (Optional) an Azure tenant id in UUID4 form.

* `peer_resource_group` - (Optional) an Azure resource group name of the peered VPC.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `state_info` - state-specific help or error information.

* `state` - is the state of the peering connection. This property is computed by Aiven 
therefore cannot be set, only read. Where state can be one of: `APPROVED`, 
`PENDING_PEER`, `ACTIVE`, `DELETED`, `DELETED_BY_PEER`, `REJECTED_BY_PEER` and 
`INVALID_SPECIFICATION`. 

Happy path of the `state` includes the following transitions of a VPC peering connection: 
`APPROVED` -> `PENDING_PEER` -> `ACTIVE`.

**More details regarding each state:**

- `APPROVED` is the initial state after the user does a successful creation of a 
peering connection resource via Terraform. 

- `PENDING_PEER` the connection enters the `PENDING_PEER` state from `APPROVED` once the 
Aiven platform has created a connection to the specified peer successfully in the cloud, 
but the connection is not active until the user completes the setup in their cloud account. 
The steps needed in the user cloud account depend on the used cloud provider.

- `ACTIVE` stands for a VPC peering connection whose setup has been completed

- `DELETED` means a user deleted the peering connection through the Aiven Terraform provider, 
or Aiven Web Console or directly via Aiven API.

- `DELETED_BY_PEER` appears when a user deleted the VPC peering connection through their cloud 
account. That is, the user deleted the peering cloud resource in their account. There are no 
transitions from this state

- `REJECTED_BY_PEER` an AWS specific state, when VPC peering connection was in the `PENDING_PEER` state, 
and the user rejected the AWS peering connection request.

- `INVALID_SPECIFICATION` is a VPC peering connection that was in the `APPROVED` state but could not be  
successfully created because of something in the user's control, for example, the peer cloud account of VPC 
doesn't exist, overlapping IP ranges, or the Aiven cloud account doesn't have permissions to peer 
there. `state_info` field contains more details about the particular issue.

Aiven ID format when importing existing resource: `<project_name>/<VPC_UUID>/<peer_cloud_account>/<peer_vpc>`.
Aiven ID format when importing existing cross-region resource: `<project_name>/<VPC_UUID>/<peer_cloud_account>/<peer_vpc>/peer_region`.
The UUID is not directly visible in the Aiven web console.
