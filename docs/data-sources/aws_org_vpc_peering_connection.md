---
page_title: "aiven_aws_org_vpc_peering_connection Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about an AWS VPC peering connection.
---

# aiven_aws_org_vpc_peering_connection (Data Source)

Gets information about an AWS VPC peering connection.

## Example Usage

```terraform
data "aiven_aws_org_vpc_peering_connection" "example" {
  organization_id     = "org1a23f456789"
  organization_vpc_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  aws_account_id      = "123456789012"
  aws_vpc_id          = "vpc-2f09a348"
  aws_vpc_region      = "us-east-1"

  /* COMPUTED FIELDS
  aws_vpc_peering_connection_id = "pcx-1234567890abcdef0"
  peering_connection_id         = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  state                         = "ACTIVE"
  */
}
```

## Schema

### Required

- `aws_account_id` (String) AWS account ID.
- `aws_vpc_id` (String) AWS VPC ID.
- `aws_vpc_region` (String) The AWS region of the peered VPC. For example, `eu-central-1`.
- `organization_id` (String) ID of an organization.
- `organization_vpc_id` (String) Organization VPC ID.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `aws_vpc_peering_connection_id` (String) The ID of the AWS VPC peering connection.
- `id` (String) Resource ID composed as: `organization_id/organization_vpc_id/aws_account_id/aws_vpc_id/aws_vpc_region`.
- `peering_connection_id` (String) Organization peering connection ID.
- `state` (String) State of the peering connection. The possible values are `ACTIVE`, `APPROVED`, `APPROVED_PEER_REQUESTED`, `DELETED`, `DELETED_BY_PEER`, `DELETING`, `ERROR`, `INVALID_SPECIFICATION`, `PENDING_PEER` and `REJECTED_BY_PEER`.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
