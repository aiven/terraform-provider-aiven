---
page_title: "aiven_upgrade_step Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an Aiven Upgrade Pipeline step between source and destination services. This feature is in the limited availability stage and may change without notice. To enable this feature, contact the sales team http://aiven.io/contact. Once it's enabled, set the PROVIDER_AIVEN_ENABLE_BETA environment variable to use the resource.
---

# aiven_upgrade_step (Resource)

Creates and manages an Aiven Upgrade Pipeline step between source and destination services. This feature is in the limited availability stage and may change without notice. To enable this feature, contact the [sales team](http://aiven.io/contact). Once it's enabled, set the `PROVIDER_AIVEN_ENABLE_BETA` environment variable to use the resource.

## Example Usage

```terraform
resource "aiven_upgrade_step" "example" {
  organization_id          = "org1a23f456789" // Force new
  destination_project_name = "prod-project" // Force new
  destination_service_name = "pg-prod" // Force new
  source_project_name      = "dev-project" // Force new
  source_service_name      = "pg-dev" // Force new

  // OPTIONAL FIELDS
  auto_validation_delay_days = 1

  /* COMPUTED FIELDS
  step_id = "550e8400-e29b-41d4-a716-446655440000"
  */
}
```

## Schema

### Required

- `destination_project_name` (String) Destination project name. Changing this property forces recreation of the resource.
- `destination_service_name` (String) Destination service name. Changing this property forces recreation of the resource.
- `organization_id` (String) ID of an organization. Changing this property forces recreation of the resource.
- `source_project_name` (String) Source project name. Changing this property forces recreation of the resource.
- `source_service_name` (String) Source service name. Changing this property forces recreation of the resource.

### Optional

- `auto_validation_delay_days` (Number) Days before automatic validation (defaults to 7). Minimum value: `1`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) Resource ID composed as: `organization_id/step_id`.
- `step_id` (String) Upgrade step ID. The possible value is `550e8400-e29b-41d4-a716-446655440000`.

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
terraform import aiven_upgrade_step.example ORGANIZATION_ID/STEP_ID
```
