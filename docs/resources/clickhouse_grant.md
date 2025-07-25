---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_clickhouse_grant Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages ClickHouse grants to give users and roles privileges to a ClickHouse service.
  There are some limitations and considerations to be aware of when using this resource:
  Users cannot have the same name as roles.Global privileges cannot be granted on the database level. To grant global privileges, use database="*".To grant a privilege on all tables of a database, omit the table and only keep the database. Don't use table="*".Privileges granted on ClickHouse Named Collections are not currently managed by this resource and will be ignored. If you have grants on Named Collections managed outside of Terraform, this resource will not attempt to alter them. For an example showing how to set up Named Collection access with S3 integration, see the ClickHouse S3 Integration example https://github.com/aiven/terraform-provider-aiven/tree/main/examples/clickhouse/clickhouse_integrations/s3.Changes first revoke all grants and then reissue the remaining grants for convergence.Some grants overlap, which can cause the Aiven Terraform Provider to detect a change even if you haven't made modifications. For example, using both DELETE and ALTER DELETE together might cause this issue.
  The ClickHouse grant privileges documentation https://clickhouse.com/docs/sql-reference/statements/grant has a list of ClickHouse privileges.
---

# aiven_clickhouse_grant (Resource)

Creates and manages ClickHouse grants to give users and roles privileges to a ClickHouse service.

There are some limitations and considerations to be aware of when using this resource:
* Users cannot have the same name as roles.
* Global privileges cannot be granted on the database level. To grant global privileges, use `database="*"`.
* To grant a privilege on all tables of a database, omit the table and only keep the database. Don't use `table="*"`.
* Privileges granted on ClickHouse Named Collections are not currently managed by this resource and will be ignored. If you have grants on Named Collections managed outside of Terraform, this resource will not attempt to alter them. For an example showing how to set up Named Collection access with S3 integration, see the [ClickHouse S3 Integration example](https://github.com/aiven/terraform-provider-aiven/tree/main/examples/clickhouse/clickhouse_integrations/s3).
* Changes first revoke all grants and then reissue the remaining grants for convergence.
* Some grants overlap, which can cause the Aiven Terraform Provider to detect a change even if you haven't made modifications. For example, using both `DELETE` and `ALTER DELETE` together might cause this issue.
  The [ClickHouse grant privileges documentation](https://clickhouse.com/docs/sql-reference/statements/grant) has a list of ClickHouse privileges.

## Example Usage

```terraform
resource "aiven_clickhouse_role" "example_role" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_clickhouse.example_clickhouse.service_name
  role         = "example-role"
}

# Grant privileges to the example role.
resource "aiven_clickhouse_grant" "role_privileges" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_clickhouse.example_clickhouse.service_name
  role         = aiven_clickhouse_role.example_role.role

  privilege_grant {
    privilege = "INSERT"
    database  = aiven_clickhouse_database.example_db.name
    table     = "example-table"
  }

  privilege_grant {
    privilege = "SELECT"
    database  = aiven_clickhouse_database.example_db.name
  }

  # Global privileges
  privilege_grant {
    privilege = "CREATE TEMPORARY TABLE"
    database  = "*"
  }

  privilege_grant {
    privilege = "SYSTEM DROP CACHE"
    database  = "*"
  }
}

# Grant the role to the user.
resource "aiven_clickhouse_user" "example_user" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_clickhouse.example_clickhouse.service_name
  username     = "example-user"
}

resource "aiven_clickhouse_grant" "user_role_assignment" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_clickhouse.example_clickhouse.service_name
  user         = aiven_clickhouse_user.example_user.username

  role_grant {
    role = aiven_clickhouse_role.example_role.role
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `project` (String) The name of the project this resource belongs to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.
- `service_name` (String) The name of the service that this resource belongs to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.

### Optional

- `privilege_grant` (Block Set) Grant privileges. Changing this property forces recreation of the resource. (see [below for nested schema](#nestedblock--privilege_grant))
- `role` (String) The role to grant privileges or roles to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.
- `role_grant` (Block Set) Grant roles. Changing this property forces recreation of the resource. (see [below for nested schema](#nestedblock--role_grant))
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `user` (String) The user to grant privileges or roles to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--privilege_grant"></a>
### Nested Schema for `privilege_grant`

Required:

- `database` (String) The database to grant access to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.

Optional:

- `column` (String) The column to grant access to. Changing this property forces recreation of the resource.
- `privilege` (String) The privileges to grant. For example: `INSERT`, `SELECT`, `CREATE TABLE`. A complete list is available in the [ClickHouse documentation](https://clickhouse.com/docs/en/sql-reference/statements/grant). Changing this property forces recreation of the resource.
- `table` (String) The table to grant access to. Changing this property forces recreation of the resource.
- `with_grant` (Boolean) Allow grantees to grant their privileges to other grantees. Changing this property forces recreation of the resource.


<a id="nestedblock--role_grant"></a>
### Nested Schema for `role_grant`

Optional:

- `role` (String) The roles to grant. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.


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
terraform import aiven_clickhouse_grant.example_grant PROJECT/SERVICE_NAME/ID
```
