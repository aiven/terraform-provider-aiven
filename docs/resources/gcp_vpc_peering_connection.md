---
page_title: "aiven_gcp_vpc_peering_connection Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages a Google Cloud VPC peering connection. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.
---

# aiven_gcp_vpc_peering_connection (Resource)

Creates and manages a Google Cloud VPC peering connection. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.

## Example Usage

```terraform
resource "aiven_gcp_vpc_peering_connection" "example" {
  gcp_project_id = "my-gcp-project" // Force new
  vpc_id         = "example-project/example-vpc" // Force new
  peer_vpc       = "my-vpc" // Force new

  /* COMPUTED FIELDS
  id        = "example-project/example-vpc/my-gcp-project/my-vpc"
  self_link = "https://www.googleapis.com/compute/v1/projects/my-gcp-project/global/networks/my-vpc"
  state     = "ACTIVE"
  state_info = {
    foo = "foo"
  }
  */
}
```

## Schema

### Required

- `gcp_project_id` (String) Google Cloud project ID. Maximum length: `1024`. Changing this property forces recreation of the resource.
- `peer_vpc` (String) Google Cloud VPC network name. Maximum length: `1024`. Changing this property forces recreation of the resource.
- `vpc_id` (String) The VPC the peering connection belongs to, in the `PROJECT/VPC_ID` format. Must match pattern: `^[^/]*/[^/]*$`. Changing this property forces recreation of the resource.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) Resource ID in the `PROJECT/VPC_ID/GCP_PROJECT_ID/PEER_VPC` format.
- `self_link` (String) Computed Google Cloud network peering link.
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
terraform import aiven_gcp_vpc_peering_connection.example PROJECT/VPC_ID/GCP_PROJECT_ID/PEER_VPC
```
