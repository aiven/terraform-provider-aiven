---
page_title: "aiven_organization_vpc Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about an existing VPC in an Aiven organization.
---

# aiven_organization_vpc (Data Source)

Gets information about an existing VPC in an Aiven organization.

## Example Usage

```terraform
data "aiven_organization_vpc" "example" {
  organization_id     = "org1a23f456789"
  organization_vpc_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"

  /* COMPUTED FIELDS
  cloud_name   = "aws-eu-west-1"
  create_time  = "2021-01-01T00:00:00Z"
  display_name = "My organization VPC"
  network_cidr = "10.0.0.0/24"
  state        = "ACTIVE"
  update_time  = "2021-01-01T00:00:00Z"
  */
}
```

## Schema

### Required

- `organization_id` (String) ID of an organization.
- `organization_vpc_id` (String) The ID of the Aiven Organization VPC.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `cloud_name` (String) The cloud provider and region where the service is hosted in the format `CLOUD_PROVIDER-REGION_NAME`. For example, `google-europe-west1` or `aws-us-east-2`.
- `create_time` (String) VPC creation timestamp.
- `display_name` (String) User defined display name for this VPC.
- `id` (String) Resource ID composed as: `organization_id/organization_vpc_id`.
- `network_cidr` (String) Network address range used by the VPC. For example, `192.168.0.0/24`.
- `state` (String) State of the VPC. The possible values are `ACTIVE`, `APPROVED`, `DELETED` and `DELETING`.
- `update_time` (String) Timestamp of last change to VPC.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
