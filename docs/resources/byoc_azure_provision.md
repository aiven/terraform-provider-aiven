---
page_title: "aiven_byoc_azure_provision Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Provisions a BYOC custom cloud environment by handing Aiven the Azure subscription and Entra ID tenant of the customer. Transitions the environment from draft to active so services can be deployed into it. Apply aiven_byoc_azure_entity first, then the customer-side Azure infrastructure (Aiven CCE enterprise application, role assignments, networking) built from its outputs, and create this resource last. Terraform orders only what this resource references: add the customer-side infrastructure to depends_on, e.g., depends_on = [azurerm_role_assignment.example], so it's created before and destroyed after this resource. The example leaves out the customer-side infrastructure, so it doesn't work on its own. Provisioning can take longer than the create timeout. If it times out, apply again: it resumes waiting. If it fails with creation_failed, fix the cause and replace the entity, e.g., terraform apply -replace=aiven_byoc_azure_entity.example, which replaces this resource too. Replace it as well to change azure_subscription_id or azure_tenant_id after provisioning. The plan can only reject a change it knows: a value unknown until apply, e.g., from a resource replaced in the same apply, updates the environment. Destroying this resource deletes the environment, after provisioning finishes if it is still running. If a refresh can't find this resource, it's removed from the state and a new create plan is generated. Beta resource. This resource is in the beta stage and may change without notice. Set the PROVIDER_AIVEN_ENABLE_BETA environment variable to use the resource.
---

# aiven_byoc_azure_provision (Resource)

Provisions a BYOC custom cloud environment by handing Aiven the Azure subscription and Entra ID tenant of the customer. Transitions the environment from `draft` to `active` so services can be deployed into it.

Apply `aiven_byoc_azure_entity` first, then the customer-side Azure infrastructure (Aiven CCE enterprise application, role assignments, networking) built from its outputs, and create this resource last. Terraform orders only what this resource references: add the customer-side infrastructure to `depends_on`, e.g., `depends_on = [azurerm_role_assignment.example]`, so it's created before and destroyed after this resource. The example leaves out the customer-side infrastructure, so it doesn't work on its own.

Provisioning can take longer than the create timeout. If it times out, apply again: it resumes waiting. If it fails with `creation_failed`, fix the cause and replace the entity, e.g., `terraform apply -replace=aiven_byoc_azure_entity.example`, which replaces this resource too. Replace it as well to change `azure_subscription_id` or `azure_tenant_id` after provisioning. The plan can only reject a change it knows: a value unknown until apply, e.g., from a resource replaced in the same apply, updates the environment.

Destroying this resource deletes the environment, after provisioning finishes if it is still running. If a refresh can't find this resource, it's removed from the state and a new create plan is generated.

~> **Beta resource**
This resource is in the beta stage and may change without notice. Set the `PROVIDER_AIVEN_ENABLE_BETA` environment variable to use the resource.

## Example Usage

```terraform
resource "aiven_byoc_azure_provision" "example" {
  organization_id             = data.aiven_organization.example.id                          // e.g. "org1a23f456789". Force new
  custom_cloud_environment_id = aiven_byoc_azure_entity.example.custom_cloud_environment_id // e.g. "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d". Force new
  azure_subscription_id       = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  azure_tenant_id             = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"

  /* COMPUTED FIELDS
  custom_cloud_names = ["foo"]
  state              = "active"
  */
}

resource "aiven_byoc_azure_entity" "example" {
  organization_id  = data.aiven_organization.example.id // e.g. "org1a23f456789". Force new
  cloud_region     = "westeurope"                       // Force new
  deployment_model = "standard"                         // Force new
  display_name     = "byoc-cloud-prod-westeurope"
  reserved_cidr    = "192.168.6.0/24" // Force new
}

data "aiven_organization" "example" {
  name = "Aiven Ltd"
}
```

## Schema

### Required

- `azure_subscription_id` (String) UUID identifying the customer's Azure subscription where BYOC infrastructure is deployed.
- `azure_tenant_id` (String) Entra ID tenant ID of the customer's directory where the Aiven CCE enterprise application is installed. Must be a valid UUID.
- `custom_cloud_environment_id` (String) ID of a custom cloud environment. Must be a valid UUID. Changing this property forces recreation of the resource.
- `organization_id` (String) ID of an organization. Changing this property forces recreation of the resource.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

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
terraform import aiven_byoc_azure_provision.example ORGANIZATION_ID/CUSTOM_CLOUD_ENVIRONMENT_ID
```

With Terraform 1.12 or later, you can also import using resource identity:

```terraform
import {
  to = aiven_byoc_azure_provision.example
  identity = {
    organization_id             = "ORGANIZATION_ID"
    custom_cloud_environment_id = "CUSTOM_CLOUD_ENVIRONMENT_ID"
  }
}
```
