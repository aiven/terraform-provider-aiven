---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_gcp_vpc_peering_connection Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  The GCP VPC Peering Connection data source provides information about the existing Aiven VPC Peering Connection.
---

# aiven_gcp_vpc_peering_connection (Data Source)

The GCP VPC Peering Connection data source provides information about the existing Aiven VPC Peering Connection.

## Example Usage

```terraform
data "aiven_gcp_vpc_peering_connection" "main" {
  vpc_id         = data.aiven_project_vpc.vpc.id
  gcp_project_id = "example-project"
  peer_vpc       = "example-network"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `gcp_project_id` (String) Google Cloud project ID. Changing this property forces recreation of the resource.
- `peer_vpc` (String) Google Cloud VPC network name. Changing this property forces recreation of the resource.
- `vpc_id` (String) The VPC the peering connection belongs to. Changing this property forces recreation of the resource.

### Read-Only

- `id` (String) The ID of this resource.
- `self_link` (String) Computed Google Cloud network peering link.
- `state` (String) State of the peering connection.
- `state_info` (Map of String) State-specific help or error information.
