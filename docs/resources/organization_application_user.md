---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_organization_application_user Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an organization application user. Application users https://aiven.io/docs/platform/howto/manage-application-users can be used for programmatic access to the platform.
  This resource is in the limited availability stage and may change without notice.  To enable this feature, contact the sales team mailto:sales@aiven.io. After it's enabled, set
  the PROVIDER_AIVEN_ENABLE_BETA environment variable to use the resource.
---

# aiven_organization_application_user (Resource)

Creates and manages an organization application user. [Application users](https://aiven.io/docs/platform/howto/manage-application-users) can be used for programmatic access to the platform. 

**This resource is in the limited availability stage and may change without notice.**  To enable this feature, contact the [sales team](mailto:sales@aiven.io). After it's enabled, set
the `PROVIDER_AIVEN_ENABLE_BETA` environment variable to use the resource.

## Example Usage

```terraform
resource "aiven_organization_application_user" "tf_user" {
  organization_id = aiven_organization.main.id
  name = "app-terraform"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the application user.
- `organization_id` (String) The ID of the organization the application user belongs to.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `email` (String) An email address automatically generated by Aiven to help identify the application user. 
				No notifications are sent to this email.
- `id` (String) A compound identifier of the resource in the format `organization_id/user_id`.
- `user_id` (String) The ID of the application user.

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
terraform import aiven_organization_application_user.example ORGANIZATION_ID/USER_ID
```
