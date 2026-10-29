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
  aws_vpc_region = "eu-west-1"

  // OPTIONAL FIELDS
  project_vpc_id = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  project        = "example-project"

  /* COMPUTED FIELDS
  aws_vpc_peering_connection_id = "pcx-0123456789abcdef0"
  id                            = "example-project/1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d/123456789012/vpc-0123456789abcdef0/eu-west-1"
  create_time                   = "2021-01-01T00:00:00Z"
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
- `aws_vpc_region` (String) The AWS region of the peered VPC. This argument must be nonempty. When upgrading a configuration that uses `""`, set the existing connection's `aws_vpc_region` from Terraform state.

### Optional

- `project` (String) The name of the project the Aiven VPC belongs to. The field is required with `project_vpc_id`.
- `project_vpc_id` (String) The ID of the Aiven VPC. The field is required with `project`. The field conflicts with `vpc_id`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `vpc_id` (String) The ID of the Aiven VPC, in the `PROJECT/PROJECT_VPC_ID` format. Exactly one of the fields must be specified: `project` or `vpc_id`.

### Read-Only

- `aws_vpc_peering_connection_id` (String) The ID of the AWS VPC peering connection.
- `create_time` (String) VPC peering connection creation timestamp. Re-creating a connection updates this timestamp even when its ID stays the same. May be `null` for legacy connections whose creation time is unknown.
- `id` (String) Resource ID in the `PROJECT/PROJECT_VPC_ID/AWS_ACCOUNT_ID/AWS_VPC_ID/AWS_VPC_REGION` format.
- `state` (String) Project VPC peering connection state. The possible values are `ACTIVE`, `APPROVED`, `APPROVED_PEER_REQUESTED`, `DELETED`, `DELETED_BY_PEER`, `DELETING`, `ERROR`, `INVALID_SPECIFICATION`, `PENDING_PEER` and `REJECTED_BY_PEER`.
- `state_info` (Map of String) State-specific help or error information.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
