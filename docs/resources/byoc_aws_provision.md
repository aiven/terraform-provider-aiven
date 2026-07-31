---
page_title: "aiven_byoc_aws_provision Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Provisions a BYOC custom cloud environment by handing Aiven the IAM role ARN created in the customer AWS account. Transitions the environment from draft to active so services can be deployed into it.
  Create this resource after the customer-side AWS infrastructure (IAM role, VPC, subnets, security groups, buckets) has been defined.
  This resource is in the beta stage and may change without notice. Set
  the PROVIDER_AIVEN_ENABLE_BETA environment variable to use the resource.
---

# aiven_byoc_aws_provision (Resource)

Provisions a BYOC custom cloud environment by handing Aiven the IAM role ARN created in the customer AWS account. Transitions the environment from `draft` to `active` so services can be deployed into it.

Create this resource after the customer-side AWS infrastructure (IAM role, VPC, subnets, security groups, buckets) has been defined.

**This resource is in the beta stage and may change without notice.** Set
the `PROVIDER_AIVEN_ENABLE_BETA` environment variable to use the resource.

## Example Usage

```terraform
resource "aiven_byoc_aws_provision" "example" {
  organization_id             = "org1a23f456789" // Force new
  custom_cloud_environment_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d" // Force new
  aws_iam_role_arn            = "arn:aws:iam::012345678901:root" // Force new

  // OPTIONAL FIELDS
  azure_client_id       = "12345678-1234-1234-1234-123456789abc" // Force new
  azure_subscription_id = "12345678-1234-1234-1234-123456789abc" // Force new
  azure_tenant_id       = "adcf7194-d877-4505-a47a-91fefd96e3b8" // Force new
  azure_client_secret   = "s3cr3t~EXAMPLE_client_secret_value" // Force new

  /* COMPUTED FIELDS
  aiven_aws_assume_role_external_id = "admin"
  aiven_aws_account_principal       = "foo"
  custom_cloud_names                = ["foo"]
  state                             = "active"
  */
}
```

## Schema

### Required

- `aws_iam_role_arn` (String) Amazon Resource Name. Maximum length: `2048`. Changing this property forces recreation of the resource.
- `custom_cloud_environment_id` (String) ID of a custom cloud environment. Length must be exactly `36`. Changing this property forces recreation of the resource.
- `organization_id` (String) ID of an organization. Changing this property forces recreation of the resource.

### Optional

- `azure_client_id` (String) Application (client) ID of the operator service principal created by Terraform. Maximum length: `36`. Changing this property forces recreation of the resource.
- `azure_client_secret` (String) Client secret of the operator service principal created by Terraform. Maximum length: `256`. Changing this property forces recreation of the resource.
- `azure_subscription_id` (String) UUID identifying the customer's Azure subscription where BYOC infrastructure is deployed. Maximum length: `36`. Changing this property forces recreation of the resource.
- `azure_tenant_id` (String) Azure tenant id in UUID4 form. Maximum length: `1024`. Changing this property forces recreation of the resource.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `aiven_aws_account_principal` (String) Entity that assumes the IAM role for controlling the BYOC account.
- `aiven_aws_assume_role_external_id` (String) External ID for assuming the IAM role for controlling the BYOC account.
- `custom_cloud_names` (Set of String) Cloud names that can be used to provision a service on this BYOC.
- `id` (String) Resource ID composed as: `organization_id/custom_cloud_environment_id`.
- `state` (String) State of this BYOC cloud. The possible values are `active`, `creating`, `creation_failed`, `deleted`, `deleting`, `deletion_failed`, `disconnected`, `draft`, `reconnecting` and `validating`.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

## Import

Import is supported using the following syntax:

```shell
terraform import aiven_byoc_aws_provision.example ORGANIZATION_ID/CUSTOM_CLOUD_ENVIRONMENT_ID
```
