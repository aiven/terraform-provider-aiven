---
page_title: "aiven_byoc_permissions Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Manages permissions for a BYOC custom cloud environment, which accounts and projects can use the cloud to deploy services. Note! The accounts and projects lists are authoritative. This differs from the avn byoc cloud permissions add, which merges into the existing set. Create this resource after the entity and provision resources (aiven_byoc_aws_entity and aiven_byoc_aws_provision, or aiven_byoc_azure_entity and aiven_byoc_azure_provision) so the custom cloud environment is active before permissions are granted. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated. Beta resource. This resource is in the beta stage and may change without notice. Set the PROVIDER_AIVEN_ENABLE_BETA environment variable to use the resource.
---

# aiven_byoc_permissions (Resource)

Manages permissions for a BYOC custom cloud environment, which accounts and projects can use the cloud to deploy services.

Note! The `accounts` and `projects` lists are authoritative. This differs from the `avn byoc cloud permissions add`, which merges into the existing set.

Create this resource after the entity and provision resources (`aiven_byoc_aws_entity` and `aiven_byoc_aws_provision`, or `aiven_byoc_azure_entity` and `aiven_byoc_azure_provision`) so the custom cloud environment is active before permissions are granted. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.

~> **Beta resource**
This resource is in the beta stage and may change without notice. Set the `PROVIDER_AIVEN_ENABLE_BETA` environment variable to use the resource.

## Example Usage

```terraform
resource "aiven_byoc_permissions" "example" {
  organization_id             = aiven_organization.example.id                                // e.g. "org1a23f456789". Force new
  custom_cloud_environment_id = aiven_byoc_aws_provision.example.custom_cloud_environment_id // e.g. "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d". Force new
  accounts                    = ["a22ba494e096", "a441b467e067"]
  projects                    = ["project-prod", "project-dev"]
}

resource "aiven_byoc_aws_provision" "example" {
  organization_id             = aiven_organization.example.id                             // e.g. "org1a23f456789". Force new
  custom_cloud_environment_id = aiven_byoc_aws_entity.example.custom_cloud_environment_id // e.g. "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d". Force new
  aws_iam_role_arn            = "arn:aws:iam::012345678901:root"
}

resource "aiven_byoc_aws_entity" "example" {
  organization_id  = aiven_organization.example.id // e.g. "org1a23f456789". Force new
  cloud_region     = "eu-west-1"                   // Force new
  deployment_model = "standard"                    // Force new
  display_name     = "byoc-cloud-prod-eu-west-1"
  reserved_cidr    = "192.168.6.0/24" // Force new
}

resource "aiven_organization" "example" {
  name = "Aiven Ltd"
}
```

## Schema

### Required

- `accounts` (Set of String) Accounts.
- `custom_cloud_environment_id` (String) ID of a custom cloud environment. Must be a valid UUID. Changing this property forces recreation of the resource.
- `organization_id` (String) ID of an organization. Changing this property forces recreation of the resource.
- `projects` (Set of String) Projects.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) Resource ID composed as: `organization_id/custom_cloud_environment_id`.

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
terraform import aiven_byoc_permissions.example ORGANIZATION_ID/CUSTOM_CLOUD_ENVIRONMENT_ID
```
