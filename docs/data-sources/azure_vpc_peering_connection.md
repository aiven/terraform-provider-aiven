---
page_title: "aiven_azure_vpc_peering_connection Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about an Azure VPC peering connection.
---

# aiven_azure_vpc_peering_connection (Data Source)

Gets information about an Azure VPC peering connection.

## Example Usage

```terraform
data "aiven_azure_vpc_peering_connection" "example" {
  azure_subscription_id = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  peer_azure_app_id     = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  peer_azure_tenant_id  = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  peer_resource_group   = "my-resource-group"
  vnet_name             = "my-vnet"

  // OPTIONAL FIELDS
  project_vpc_id = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  project        = "example-project"

  /* COMPUTED FIELDS
  id                    = "example-project/1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d/2a3b4c5d-6e7f-4a8b-9c0d-1e2f3a4b5c6d/my-vnet"
  peering_connection_id = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  create_time           = "2021-01-01T00:00:00Z"
  state                 = "ACTIVE"
  state_info = {
    foo = "foo"
  }
  */
}
```

## Schema

### Required

- `azure_subscription_id` (String) The ID of the Azure subscription in UUID4 format.
- `peer_azure_app_id` (String) Azure app registration id in UUID4 form that is allowed to create a peering to the peer vnet.
- `peer_azure_tenant_id` (String) Azure tenant id in UUID4 form.
- `peer_resource_group` (String) Azure resource group name of the peered VPC.
- `vnet_name` (String) The name of the Azure VNet.

### Optional

- `project` (String) The name of the project the Aiven VPC belongs to. The field is required with `project_vpc_id`.
- `project_vpc_id` (String) The ID of the Aiven VPC. The field is required with `project`. The field conflicts with `vpc_id`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `vpc_id` (String) The ID of the Aiven VPC, in the `PROJECT/PROJECT_VPC_ID` format. Exactly one of the fields must be specified: `project` or `vpc_id`.

### Read-Only

- `create_time` (String) VPC peering connection creation timestamp. Re-creating a connection updates this timestamp even when its ID stays the same.
- `id` (String) Resource ID in the `PROJECT/PROJECT_VPC_ID/AZURE_SUBSCRIPTION_ID/VNET_NAME` format.
- `peering_connection_id` (String) Legacy attribute retained for state compatibility. It is not populated by the Project VPC API.
- `state` (String) Project VPC peering connection state. The possible values are `ACTIVE`, `APPROVED`, `APPROVED_PEER_REQUESTED`, `DELETED`, `DELETED_BY_PEER`, `DELETING`, `ERROR`, `INVALID_SPECIFICATION`, `PENDING_PEER` and `REJECTED_BY_PEER`.
- `state_info` (Map of String) State-specific help or error information.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
