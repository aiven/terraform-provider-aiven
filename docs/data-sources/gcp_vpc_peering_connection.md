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
  vpc_id         = "example-project/example-vpc"
  peer_vpc       = "my-vpc"

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

- `gcp_project_id` (String) Google Cloud project ID.
- `peer_vpc` (String) Google Cloud VPC network name.
- `vpc_id` (String) The VPC the peering connection belongs to, in the `PROJECT/VPC_ID` format.

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

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
