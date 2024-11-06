---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_organization_user_list Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  List of users of the organization.
  This resource is in the beta stage and may change without notice. Set
  the PROVIDER_AIVEN_ENABLE_BETA environment variable to use the resource.
---

# aiven_organization_user_list (Data Source)

List of users of the organization. 

**This resource is in the beta stage and may change without notice.** Set
the `PROVIDER_AIVEN_ENABLE_BETA` environment variable to use the resource.



<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `id` (String) Organization id. Example: `org12345678`.
- `name` (String) Organization name. Example: `aiven`.

### Read-Only

- `users` (List of Object) List of users of the organization (see [below for nested schema](#nestedatt--users))

<a id="nestedatt--users"></a>
### Nested Schema for `users`

Read-Only:

- `is_super_admin` (Boolean)
- `join_time` (String)
- `last_activity_time` (String)
- `user_id` (String)
- `user_info` (List of Object) (see [below for nested schema](#nestedobjatt--users--user_info))

<a id="nestedobjatt--users--user_info"></a>
### Nested Schema for `users.user_info`

Read-Only:

- `city` (String)
- `country` (String)
- `create_time` (String)
- `department` (String)
- `is_application_user` (Boolean)
- `job_title` (String)
- `managed_by_scim` (Boolean)
- `managing_organization_id` (String)
- `real_name` (String)
- `state` (String)
- `user_email` (String)