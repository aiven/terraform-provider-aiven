---
page_title: "aiven_azure_privatelink Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an Azure Private Link for selected Aiven services https://aiven.io/docs/platform/howto/use-azure-privatelink in a VPC. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.
---

# aiven_azure_privatelink (Resource)

Creates and manages an Azure Private Link for [selected Aiven services](https://aiven.io/docs/platform/howto/use-azure-privatelink) in a VPC. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.

## Example Usage

```terraform
resource "aiven_azure_privatelink" "example" {
  project               = "my-project" // Force new
  service_name          = "foo" // Force new
  user_subscription_ids = ["adcf7194-d877-4505-a47a-91fefd96e3b8"]

  /* COMPUTED FIELDS
  azure_service_id    = "foo"
  azure_service_alias = "foo"
  state               = "active"
  */
}
```

## Schema

### Required

- `project` (String) Project name. Changing this property forces recreation of the resource.
- `service_name` (String) Service name. Changing this property forces recreation of the resource.
- `user_subscription_ids` (Set of String) IDs of Azure subscriptions allowed to connect to the service.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `azure_service_alias` (String) Azure Privatelink service alias.
- `azure_service_id` (String) Azure Privatelink service ID.
- `id` (String) Resource ID composed as: `project/service_name`.
- `message` (String, Deprecated) Legacy response message retained for backward compatibility. **Deprecated**: This attribute is retained only for compatibility with state created by older provider versions and is no longer populated.
- `state` (String) Privatelink resource state. The possible values are `active`, `creating` and `deleting`.

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
terraform import aiven_azure_privatelink.example PROJECT/SERVICE_NAME
```
