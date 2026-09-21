---
page_title: "aiven_aws_vpc_peering_connection Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about an AWS VPC peering connection.
---

# aiven_aws_vpc_peering_connection (Data Source)

Gets information about an AWS VPC peering connection.

## Example Usage

```terraform
data "aiven_aws_vpc_peering_connection" "example" {
  aws_account_id = "123456789012"
  aws_vpc_id     = "vpc-0123456789abcdef0"
  vpc_id         = "example-project/example-vpc"
  aws_vpc_region = "eu-west-1"

  /* COMPUTED FIELDS
  aws_vpc_peering_connection_id = "pcx-0123456789abcdef0"
  id                            = "example-project/example-vpc/123456789012/vpc-0123456789abcdef0/eu-west-1"
  state                         = "ACTIVE"
  state_info = {
    foo = "foo"
  }
  */
}
```

## Schema

### Required

- `aws_account_id` (String) AWS account ID.
- `aws_vpc_id` (String) AWS VPC ID.
- `aws_vpc_region` (String) The AWS region of the peered VPC. An empty string uses the Aiven VPC region on creation and retains the existing connection on subsequent changes. Changing to a different non-empty region forces recreation. The configured empty string is retained in state.
- `vpc_id` (String) The ID of the Aiven VPC, in the `PROJECT/VPC_ID` format.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `aws_vpc_peering_connection_id` (String) The ID of the AWS VPC peering connection.
- `id` (String) Resource ID in the `PROJECT/VPC_ID/AWS_ACCOUNT_ID/AWS_VPC_ID/AWS_VPC_REGION` format.
- `state` (String) Project VPC peering connection state. The possible values are `ACTIVE`, `APPROVED`, `APPROVED_PEER_REQUESTED`, `DELETED`, `DELETED_BY_PEER`, `DELETING`, `ERROR`, `INVALID_SPECIFICATION`, `PENDING_PEER` and `REJECTED_BY_PEER`.
- `state_info` (Map of String) State-specific help or error information.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
