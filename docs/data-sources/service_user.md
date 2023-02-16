---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_service_user Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  The Service User data source provides information about the existing Aiven Service User.
---

# aiven_service_user (Data Source)

The Service User data source provides information about the existing Aiven Service User.

## Example Usage

```terraform
data "aiven_service_user" "myserviceuser" {
  project      = aiven_project.myproject.project
  service_name = aiven_pg.mypg.service_name
  username     = "<USERNAME>"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `project` (String) Identifies the project this resource belongs to. To set up proper dependencies please refer to this variable as a reference. This property cannot be changed, doing so forces recreation of the resource.
- `service_name` (String) Specifies the name of the service that this resource belongs to. To set up proper dependencies please refer to this variable as a reference. This property cannot be changed, doing so forces recreation of the resource.
- `username` (String) The actual name of the service user. To set up proper dependencies please refer to this variable as a reference. This property cannot be changed, doing so forces recreation of the resource.

### Read-Only

- `access_cert` (String, Sensitive) Access certificate for the user if applicable for the service in question
- `access_key` (String, Sensitive) Access certificate key for the user if applicable for the service in question
- `authentication` (String) Authentication details. The possible values are `caching_sha2_password` and `mysql_native_password`.
- `id` (String) The ID of this resource.
- `password` (String, Sensitive) The password of the service user ( not applicable for all services ).
- `pg_allow_replication` (Boolean) Postgres specific field, defines whether replication is allowed. This property cannot be changed, doing so forces recreation of the resource.
- `redis_acl_categories` (List of String) Redis specific field, defines command category rules. The field is required with`redis_acl_commands` and `redis_acl_keys`. This property cannot be changed, doing so forces recreation of the resource.
- `redis_acl_channels` (List of String) Redis specific field, defines the permitted pub/sub channel patterns. This property cannot be changed, doing so forces recreation of the resource.
- `redis_acl_commands` (List of String) Redis specific field, defines rules for individual commands. The field is required with`redis_acl_categories` and `redis_acl_keys`. This property cannot be changed, doing so forces recreation of the resource.
- `redis_acl_keys` (List of String) Redis specific field, defines key access rules. The field is required with`redis_acl_categories` and `redis_acl_keys`. This property cannot be changed, doing so forces recreation of the resource.
- `type` (String) Type of the user account. Tells wether the user is the primary account or a regular account.

