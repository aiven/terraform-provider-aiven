---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_mysql_user Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an Aiven for MySQL® service user.
---

# aiven_mysql_user (Resource)

Creates and manages an Aiven for MySQL® service user.

## Example Usage

```terraform
resource "aiven_mysql_user" "example_mysql_user" {
  service_name = aiven_mysql.example_mysql.service_name
  project      = data.aiven_project.example_project.project
  username     = "example-mysql-user"
  password     = var.service_user_pw
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `project` (String) The name of the project this resource belongs to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.
- `service_name` (String) The name of the service that this resource belongs to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.
- `username` (String) The name of the MySQL service user. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.

### Optional

- `authentication` (String) Authentication details. The possible values are `caching_sha2_password`, `mysql_native_password` and `null`.
- `password` (String, Sensitive) The password of the MySQL service user.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `access_cert` (String, Sensitive) Access certificate for the user.
- `access_key` (String, Sensitive) Access certificate key for the user.
- `id` (String) The ID of this resource.
- `type` (String) User account type, such as primary or regular account.

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
terraform import aiven_mysql_user.foo PROJECT/SERVICE_NAME/USERNAME
```
