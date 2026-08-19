---
page_title: "aiven_gcp_org_vpc_peering_connection Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  The GCP VPC Peering Connection data source provides information about the existing Aiven VPC Peering Connection.
---

# aiven_gcp_org_vpc_peering_connection (Data Source)

The GCP VPC Peering Connection data source provides information about the existing Aiven VPC Peering Connection.

## Example Usage

```terraform
data "aiven_gcp_org_vpc_peering_connection" "example" {
  organization_id     = "org1a23f456789"
  organization_vpc_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  gcp_project_id      = "my-gcp-project"
  peer_vpc            = "my-vpc"

  /* COMPUTED FIELDS
  self_link = "https://www.googleapis.com/compute/v1/projects/my-gcp-project/global/networks/my-vpc"
  state     = "ACTIVE"
  */
}
```

## Schema

### Required

- `gcp_project_id` (String) Google Cloud project ID.
- `organization_id` (String) ID of an organization.
- `organization_vpc_id` (String) Organization VPC ID.
- `peer_vpc` (String) Google Cloud VPC network name.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) Resource ID composed as: `organization_id/organization_vpc_id/gcp_project_id/peer_vpc`.
- `self_link` (String) Computed Google Cloud network peering link.
- `state` (String) State of the peering connection. The possible values are `ACTIVE`, `APPROVED`, `APPROVED_PEER_REQUESTED`, `DELETED`, `DELETED_BY_PEER`, `DELETING`, `ERROR`, `INVALID_SPECIFICATION`, `PENDING_PEER` and `REJECTED_BY_PEER`.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
