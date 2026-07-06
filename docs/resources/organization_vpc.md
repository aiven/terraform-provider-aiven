---
page_title: "aiven_organization_vpc Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages a VPC for an Aiven organization. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.
---

# aiven_organization_vpc (Resource)

Creates and manages a VPC for an Aiven organization. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.

## Example Usage

```terraform
resource "aiven_organization_vpc" "example" {
  organization_id = "org1a23f456789" // Force new
  cloud_name      = "aws-eu-west-1" // Force new
  network_cidr    = "10.0.0.0/24" // Force new

  // OPTIONAL FIELDS
  display_name = "My organization VPC"

  /* COMPUTED FIELDS
  organization_vpc_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  create_time         = "2021-01-01T00:00:00Z"
  state               = "ACTIVE"
  update_time         = "2021-01-01T00:00:00Z"
  */
}
```

## Schema

### Required

- `cloud_name` (String) The cloud provider and region where the service is hosted in the format `CLOUD_PROVIDER-REGION_NAME`. For example, `google-europe-west1` or `aws-us-east-2`. Changing this property forces recreation of the resource.
- `network_cidr` (String) Network address range used by the VPC. For example, `192.168.0.0/24`. Changing this property forces recreation of the resource.
- `organization_id` (String) ID of an organization. Maximum length: `36`. Changing this property forces recreation of the resource.

### Optional

- `display_name` (String) User defined display name for this VPC. Maximum length: `64`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `create_time` (String) VPC creation timestamp.
- `id` (String) Resource ID composed as: `organization_id/organization_vpc_id`.
- `organization_vpc_id` (String) The ID of the Aiven Organization VPC.
- `state` (String) State of the VPC. The possible values are `ACTIVE`, `APPROVED`, `DELETED` and `DELETING`.
- `update_time` (String) Timestamp of last change to VPC.

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
terraform import aiven_organization_vpc.example ORGANIZATION_ID/ORGANIZATION_VPC_ID
```
