---
page_title: "aiven_gcp_vpc_peering_connection Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages a Google Cloud VPC peering connection. Configure either vpc_id or both project and project_vpc_id to select the Aiven VPC. If creation succeeds but the readiness check fails, Terraform keeps the connection in state for replacement. To retry with the same peer identifiers, use destroy-before-create for the recovery apply. With create_before_destroy, the existing connection prevents creation of its replacement. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.
---

# aiven_gcp_vpc_peering_connection (Resource)

Creates and manages a Google Cloud VPC peering connection. Configure either `vpc_id` or both `project` and `project_vpc_id` to select the Aiven VPC.

If creation succeeds but the readiness check fails, Terraform keeps the connection in state for replacement. To retry with the same peer identifiers, use destroy-before-create for the recovery apply. With `create_before_destroy`, the existing connection prevents creation of its replacement. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.

## Example Usage

```terraform
resource "aiven_gcp_vpc_peering_connection" "example" {
  gcp_project_id = "my-gcp-project" // Force new
  peer_vpc       = "my-vpc" // Force new

  // OPTIONAL FIELDS
  project_vpc_id = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  project        = "example-project"

  /* COMPUTED FIELDS
  id          = "example-project/1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d/my-gcp-project/my-vpc"
  create_time = "2021-01-01T00:00:00Z"
  self_link   = "https://www.googleapis.com/compute/v1/projects/my-gcp-project/global/networks/my-vpc"
  state       = "ACTIVE"
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

### Optional

- `project` (String) The name of the project the Aiven VPC belongs to. Changing the Aiven VPC forces recreation of the resource. The field is required with `project_vpc_id`.
- `project_vpc_id` (String) The ID of the Aiven VPC. Changing the Aiven VPC forces recreation of the resource. The field is required with `project`. The field conflicts with `vpc_id`. Must be a valid UUID. Must match pattern: `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `vpc_id` (String) The VPC the peering connection belongs to, in the `PROJECT/PROJECT_VPC_ID` format. Changing the Aiven VPC forces recreation of the resource; switching to `project` and `project_vpc_id` doesn't. Exactly one of the fields must be specified: `project` or `vpc_id`. Must match pattern: `^[^/]*/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`.

### Read-Only

- `create_time` (String) VPC peering connection creation timestamp. Re-creating a connection updates this timestamp even when its ID stays the same. May be `null` for legacy connections whose creation time is unknown.
- `id` (String) Resource ID in the `PROJECT/PROJECT_VPC_ID/GCP_PROJECT_ID/PEER_VPC` format.
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
terraform import aiven_gcp_vpc_peering_connection.example PROJECT/PROJECT_VPC_ID/GCP_PROJECT_ID/PEER_VPC
```
