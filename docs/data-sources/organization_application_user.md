---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_organization_application_user Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about an application user.
  This data source is in the limited availability stage and may change without notice.  To enable this feature, contact the sales team mailto:sales@aiven.io. After it's enabled, set
  the PROVIDER_AIVEN_ENABLE_BETA environment variable to use the data source.
---

# aiven_organization_application_user (Data Source)

Gets information about an application user. 

**This data source is in the limited availability stage and may change without notice.**  To enable this feature, contact the [sales team](mailto:sales@aiven.io). After it's enabled, set
the `PROVIDER_AIVEN_ENABLE_BETA` environment variable to use the data source.

## Example Usage

```terraform
data "aiven_organization_application_user" "tf_user" {
  organization_id = aiven_organization.main.id
  user_id = "u123a456b7890c"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `organization_id` (String) The ID of the organization the application user belongs to.
- `user_id` (String) The ID of the application user.

### Read-Only

- `email` (String) The auto-generated email address of the application user.
- `name` (String) Name of the application user.
