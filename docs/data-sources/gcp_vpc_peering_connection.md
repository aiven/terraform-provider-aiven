---
page_title: "aiven_gcp_vpc_peering_connection Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  The GCP VPC Peering Connection data source provides information about the existing Aiven VPC Peering Connection.
---

# aiven_gcp_vpc_peering_connection (Data Source)

The GCP VPC Peering Connection data source provides information about the existing Aiven VPC Peering Connection.

## Example Usage

```terraform
data "aiven_gcp_vpc_peering_connection" "example" {
  gcp_project_id = "my-gcp-project"
  peer_vpc       = "my-vpc"

  // OPTIONAL FIELDS
  project        = "example-project"
  project_vpc_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"

  /* COMPUTED FIELDS
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

- `gcp_project_id` (String) Google Cloud project ID.
- `peer_vpc` (String) Google Cloud VPC network name.

### Optional

- `project` (String) The name of the project the Aiven VPC belongs to. The field is required with `project_vpc_id`.
- `project_vpc_id` (String) The ID of the Aiven VPC. The field is required with `project`. The field conflicts with `vpc_id`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `vpc_id` (String, Deprecated) The VPC the peering connection belongs to, in the `PROJECT/PROJECT_VPC_ID` format. Exactly one of the fields must be specified: `project` or `vpc_id`. **Deprecated**: Use the separate `project` and `project_vpc_id` fields instead.

### Read-Only

- `create_time` (String) VPC peering connection creation timestamp. Re-creating a connection updates this timestamp even when its ID stays the same. May be `null` for legacy connections whose creation time is unknown.
- `id` (String) Resource ID composed as: `project/project_vpc_id/gcp_project_id/peer_vpc`.
- `self_link` (String) Computed Google Cloud network peering link.
- `state` (String) Project VPC peering connection state. The possible values are `ACTIVE`, `APPROVED`, `APPROVED_PEER_REQUESTED`, `DELETED`, `DELETED_BY_PEER`, `DELETING`, `ERROR`, `INVALID_SPECIFICATION`, `PENDING_PEER` and `REJECTED_BY_PEER`.
- `state_info` (Map of String) State-specific help or error information.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
