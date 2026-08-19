---
page_title: "aiven_azure_org_vpc_peering_connection Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an Azure VPC peering connection with an Aiven VPC. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.
---

# aiven_azure_org_vpc_peering_connection (Resource)

Creates and manages an Azure VPC peering connection with an Aiven VPC. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.

## Example Usage

```terraform
resource "aiven_azure_org_vpc_peering_connection" "example" {
  organization_id       = "org1a23f456789" // Force new
  organization_vpc_id   = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d" // Force new
  azure_subscription_id = "12345678-1234-1234-1234-123456789012" // Force new
  vnet_name             = "my-vnet" // Force new
  peer_resource_group   = "my-resource-group" // Force new
  peer_azure_app_id     = "87654321-4321-4321-4321-210987654321" // Force new
  peer_azure_tenant_id  = "11111111-2222-3333-4444-555555555555" // Force new

  /* COMPUTED FIELDS
  peering_connection_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  state                 = "ACTIVE"
  */
}
```

## Schema

### Required

- `azure_subscription_id` (String) The ID of the Azure subscription in UUID4 format. Maximum length: `1024`. Changing this property forces recreation of the resource.
- `organization_id` (String) ID of an organization. Changing this property forces recreation of the resource.
- `organization_vpc_id` (String) Organization VPC ID. Changing this property forces recreation of the resource.
- `peer_azure_app_id` (String) The ID of the Azure app that is allowed to create a peering to the Azure Virtual Network (VNet) in UUID4 format. Maximum length: `1024`. Changing this property forces recreation of the resource.
- `peer_azure_tenant_id` (String) The Azure tenant ID in UUID4 format. Maximum length: `1024`. Changing this property forces recreation of the resource.
- `peer_resource_group` (String) The name of the Azure resource group associated with the VNet. Maximum length: `1024`. Changing this property forces recreation of the resource.
- `vnet_name` (String) The name of the Azure VNet. Maximum length: `1024`. Changing this property forces recreation of the resource.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) Resource ID composed as: `organization_id/organization_vpc_id/azure_subscription_id/vnet_name/peer_resource_group`.
- `peering_connection_id` (String) Organization peering connection ID.
- `state` (String) State of the peering connection. The possible values are `ACTIVE`, `APPROVED`, `APPROVED_PEER_REQUESTED`, `DELETED`, `DELETED_BY_PEER`, `DELETING`, `ERROR`, `INVALID_SPECIFICATION`, `PENDING_PEER` and `REJECTED_BY_PEER`.

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
terraform import aiven_azure_org_vpc_peering_connection.example ORGANIZATION_ID/ORGANIZATION_VPC_ID/AZURE_SUBSCRIPTION_ID/VNET_NAME/PEER_RESOURCE_GROUP
```
