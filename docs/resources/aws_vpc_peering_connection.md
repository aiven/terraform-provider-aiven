---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_aws_vpc_peering_connection Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an AWS VPC peering connection with an Aiven VPC.
---

# aiven_aws_vpc_peering_connection (Resource)

Creates and manages an AWS VPC peering connection with an Aiven VPC.

## Example Usage

```terraform
resource "aiven_project_vpc" "example_vpc" {
  project      = data.aiven_project.example_project.project
  cloud_name   = "aws-us-east-2"
  network_cidr = "192.168.1.0/24"
}


resource "aiven_aws_vpc_peering_connection" "aws_to_aiven_peering" {
  vpc_id         = aiven_project_vpc.example_vpc.id
  aws_account_id = var.aws_id
  aws_vpc_id     = "vpc-1a2b3c4d5e6f7g8h9"
  aws_vpc_region = "aws-us-east-2"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `aws_account_id` (String) AWS account ID. Changing this property forces recreation of the resource.
- `aws_vpc_id` (String) AWS VPC ID. Changing this property forces recreation of the resource.
- `aws_vpc_region` (String) The AWS region of the peered VPC, if different from the Aiven VPC region. Changing this property forces recreation of the resource.
- `vpc_id` (String) The ID of the Aiven VPC. Changing this property forces recreation of the resource.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `aws_vpc_peering_connection_id` (String) The ID of the AWS VPC peering connection.
- `id` (String) The ID of this resource.
- `state` (String) The state of the peering connection.
- `state_info` (Map of String) State-specific help or error information.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `default` (String)
- `delete` (String)
- `read` (String)
- `update` (String)

## Import

Import is supported using the following syntax:

```shell
terraform import aiven_aws_vpc_peering_connection.aws_to_aiven_peering PROJECT/VPC_ID/AWS_ACCOUNT_ID/AWS_VPC_ID/AWS_VPC_REGION
```
