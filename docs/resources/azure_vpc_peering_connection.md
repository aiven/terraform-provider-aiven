---
page_title: "aiven_azure_vpc_peering_connection Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an Azure VPC peering connection with an Aiven VPC. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.
---

# aiven_azure_vpc_peering_connection (Resource)

Creates and manages an Azure VPC peering connection with an Aiven VPC. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.

## Example Usage

```terraform
resource "aiven_azure_vpc_peering_connection" "example" {
  azure_subscription_id = "12345678-1234-1234-1234-123456789012" // Force new
  vnet_name             = "my-vnet"                              // Force new
  peer_azure_app_id     = "87654321-4321-4321-4321-210987654321" // Force new
  peer_azure_tenant_id  = "11111111-2222-3333-4444-555555555555" // Force new
  peer_resource_group   = "my-resource-group"                    // Force new

  // OPTIONAL FIELDS
  project        = aiven_project_vpc.example.project        // e.g. "example-project"
  project_vpc_id = aiven_project_vpc.example.project_vpc_id // e.g. "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"

  /* COMPUTED FIELDS
  peering_connection_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  create_time           = "2021-01-01T00:00:00Z"
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

- `azure_subscription_id` (String) The ID of the Azure subscription in UUID4 format. Maximum length: `1024`. Changing this property forces recreation of the resource.
- `peer_azure_app_id` (String) Azure app registration id in UUID4 form that is allowed to create a peering to the peer vnet. Maximum length: `1024`. Changing this property forces recreation of the resource.
- `peer_azure_tenant_id` (String) Azure tenant id in UUID4 form. Maximum length: `1024`. Changing this property forces recreation of the resource.
- `peer_resource_group` (String) Azure resource group name of the peered VPC. Maximum length: `1024`. Changing this property forces recreation of the resource.
- `vnet_name` (String) The name of the Azure VNet. Maximum length: `1024`. Changing this property forces recreation of the resource.

### Optional

- `project` (String) The name of the project the Aiven VPC belongs to. Changing the Aiven VPC forces recreation of the resource. The field is required with `project_vpc_id`.
- `project_vpc_id` (String) The ID of the Aiven VPC. Changing the Aiven VPC forces recreation of the resource. The field is required with `project`. The field conflicts with `vpc_id`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `vpc_id` (String, Deprecated) The ID of the Aiven VPC, in the `PROJECT/PROJECT_VPC_ID` format. Changing the Aiven VPC forces recreation of the resource; switching to `project` and `project_vpc_id` doesn't. Exactly one of the fields must be specified: `project` or `vpc_id`. Must match pattern: `^[^/]*/[^/]*$`. **Deprecated**: Use the separate `project` and `project_vpc_id` fields instead.

### Read-Only

- `create_time` (String) VPC peering connection creation timestamp. Re-creating a connection updates this timestamp even when its ID stays the same.
- `id` (String) Resource ID composed as: `project/project_vpc_id/azure_subscription_id/vnet_name`.
- `peering_connection_id` (String) Legacy attribute retained for state compatibility. It is not populated by the Project VPC API.
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
terraform import aiven_azure_vpc_peering_connection.example PROJECT/PROJECT_VPC_ID/AZURE_SUBSCRIPTION_ID/VNET_NAME
```
