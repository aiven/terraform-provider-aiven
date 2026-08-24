---
page_title: "aiven_azure_org_vpc_peering_connection Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about an Azure VPC peering connection.
---

# aiven_azure_org_vpc_peering_connection (Data Source)

Gets information about an Azure VPC peering connection.

## Example Usage

```terraform
data "aiven_azure_org_vpc_peering_connection" "example" {
  organization_id       = "org1a23f456789"
  organization_vpc_id   = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  azure_subscription_id = "12345678-1234-1234-1234-123456789012"
  vnet_name             = "my-vnet"
  peer_resource_group   = "my-resource-group"

  /* COMPUTED FIELDS
  peer_azure_app_id     = "87654321-4321-4321-4321-210987654321"
  peer_azure_tenant_id  = "11111111-2222-3333-4444-555555555555"
  peering_connection_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  state                 = "ACTIVE"
  */
}
```

## Schema

### Required

- `azure_subscription_id` (String) The ID of the Azure subscription in UUID4 format.
- `organization_id` (String) ID of an organization.
- `organization_vpc_id` (String) Organization VPC ID.
- `peer_resource_group` (String) The name of the Azure resource group associated with the VNet.
- `vnet_name` (String) The name of the Azure VNet.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) Resource ID composed as: `organization_id/organization_vpc_id/azure_subscription_id/vnet_name/peer_resource_group`.
- `peer_azure_app_id` (String) The ID of the Azure app that is allowed to create a peering to the Azure Virtual Network (VNet) in UUID4 format.
- `peer_azure_tenant_id` (String) The Azure tenant ID in UUID4 format.
- `peering_connection_id` (String) Organization peering connection ID.
- `state` (String) State of the peering connection. The possible values are `ACTIVE`, `APPROVED`, `APPROVED_PEER_REQUESTED`, `DELETED`, `DELETED_BY_PEER`, `DELETING`, `ERROR`, `INVALID_SPECIFICATION`, `PENDING_PEER` and `REJECTED_BY_PEER`.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
