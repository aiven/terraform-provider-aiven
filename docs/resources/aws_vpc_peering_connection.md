---
page_title: "aiven_aws_vpc_peering_connection Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an AWS VPC peering connection with an Aiven VPC. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.
---

# aiven_aws_vpc_peering_connection (Resource)

Creates and manages an AWS VPC peering connection with an Aiven VPC. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.

## Example Usage

```terraform
resource "aiven_aws_vpc_peering_connection" "example" {
  aws_account_id = "123456789012" // Force new
  aws_vpc_id     = "vpc-0123456789abcdef0" // Force new
  vpc_id         = "example-project/example-vpc" // Force new
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

- `aws_account_id` (String) AWS account ID. Maximum length: `1024`. Changing this property forces recreation of the resource.
- `aws_vpc_id` (String) AWS VPC ID. Maximum length: `1024`. Changing this property forces recreation of the resource.
- `aws_vpc_region` (String) The AWS region of the peered VPC. An empty string uses the Aiven VPC region on creation and retains the existing connection on subsequent changes. Changing to a different non-empty region forces recreation. The configured empty string is retained in state. Maximum length: `1024`.
- `vpc_id` (String) The ID of the Aiven VPC, in the `PROJECT/VPC_ID` format. Must match pattern: `^[^/]*/[^/]*$`. Changing this property forces recreation of the resource.

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

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `default` (String, Deprecated) Timeout for all operations. Deprecated, use operation-specific timeouts instead.
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

## Import

Import is supported using the following syntax:

```shell
terraform import aiven_aws_vpc_peering_connection.example PROJECT/VPC_ID/AWS_ACCOUNT_ID/AWS_VPC_ID/AWS_VPC_REGION
```
