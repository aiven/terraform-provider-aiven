---
page_title: "Migrate resources"
---

# Migrate resources

You can migrate your existing resources to new resources by updating your Terraform configuration and state file. Resources need to be migrated when:

* A resource is deprecated in favor of a replacement. Deprecation notices are shown when planning and applying changes, in the documentation, and in the changelog.
* A resource is changed in another interface like the Aiven Console or API. This can happen when Aiven makes changes like migrating customers from an
  old feature to a new one. It can also happen in cases where you can only make a change, like an upgrade, in one of these other interfaces.
  In these cases, you need to update your Terraform configuration to match the actual state.

Details about changes to resources are available in the [changelog](https://github.com/aiven/terraform-provider-aiven/blob/main/CHANGELOG.md).
Aiven also sends email notifications for situations like automatic migrations to new resources.

## Migrate deprecated resources

To replace resources that are deprecated:

1. Back up your Terraform state file, `terraform.tfstate`, so you can restore the previous state if needed.

2. Replace all instances of the deprecated resource with the new resource in your Terraform files.
   In the following example, the `aiven_database` resource was replaced with `aiven_pg_database`:

      ```hcl
      - resource "aiven_database" "mydatabase" {
          project       = "myproject"
          service_name  = "mypgservice"
          database_name = "example-database"
      }


      + resource "aiven_pg_database" "mydatabase" {
          project       = "myproject"
          service_name  = "mypgservice"
          database_name = "example-database"
      }
      ```

      -> **Tip**
      To list all resources in the state file, run: `terraform state list`.

3. Preview removing all instances of the deprecated resource from Terraform's control by running:

      ```bash
      terraform state rm -dry-run RESOURCE.RESOURCE_NAME
      ```

      For example, to preview removing the `aiven_database` resource named `mydatabase`, run:

      ```bash
      terraform state rm -dry-run aiven_database.mydatabase
      ```

4. To remove the resources, run the command without the `-dry-run` flag:

    ```bash
    terraform state rm RESOURCE.RESOURCE_NAME
    ```

5. Add the new resources to Terraform by [importing them](https://registry.terraform.io/providers/aiven/aiven/latest/docs/guides/importing-resources).

   For example, to add the `aiven_pg_database` resource, run:

    ```bash
    terraform import aiven_pg_database.mydatabase myproject/mypgservice/example-database
     ```

6. To preview your configuration changes, run:

    ```bash
    terraform plan
    ```

7. To add the new resources, apply the new configuration by running:

    ```bash
    terraform apply --auto-approve
     ```

## Migrate to `aiven_organization_permission`

The `aiven_project_user` and `aiven_organization_group_project` resources have been replaced by
[the `aiven_organization_permission` resource](https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/organization_permission).

~> **Important**
Do not use the `aiven_project_user` or `aiven_organization_group_project` resource with `aiven_organization_permission`.
All of these resources manage organization-level permissions and using them together can cause conflicts and unexpected behavior.

This example shows you how to migrate to the new permission resource. The following file has a user assigned to a project with the
read only role and a group assigned to the same project with the operator role.

```hcl
# Project
data "aiven_project" "example_project" {
  project = "example-project"
}

# Group
data "aiven_organization_user_group" "example_group" {
  name            = "example-group"
  organization_id = aiven_organization.main.id
}

# Assign a user to the project with the read only role
resource "aiven_project_user" "example_project_user" {
  project     = data.aiven_project.example_project.project
  email       = "dana@example.com"
  member_type = "read_only"
}

# Assign a group to the project with the operator role
resource "aiven_organization_group_project" "example" {
  group_id = data.aiven_organization_user_group.example_group.group_id
  project  = data.aiven_project.example_project.project
  role     = "operator"
}
```

1. Replace all `aiven_project_user` and `aiven_organization_group_project` resources with the `aiven_organization_permission` resource:

      ```hcl
      # Project
      data "aiven_project" "example_project" {
        project = "example-project"
      }

      # Group
      data "aiven_organization_user_group" "example_group" {
        name            = "example-group"
        organization_id = aiven_organization.main.id
      }

      # Organization users
      data "aiven_organization_user_list" "users" {
       id    = aiven_organization.main.id
       name  = aiven.organization.main.name
      }

      # Grant the read_only role to the user and the operator role to the group
      resource "aiven_organization_permission" "example_permissions" {
        organization_id = data.aiven_organization.main.id
        resource_id     = data.aiven_project.example_project.project
        resource_type   = "project"
        permissions {
          permissions = [
            "read_only"
          ]
          principal_id   = one([for user in data.aiven_organization_user_list.users.users : user.user_id if user.user_info[0].user_email == "izumi@example.com"])
          principal_type = "user"
        }
        permissions {
          permissions = [
            "operator"
          ]
          principal_id   = data.aiven_organization_user_group.example_group.group_id
          principal_type = "user_group"
        }
      }
      ```

2. Remove the deprecated resources from Terraform's control by running:

      ```bash
      terraform state rm aiven_organization_group_project.example
      terraform state rm aiven_project_user.example_project_user
      ```

3. Preview your changes by running:

      ```bash
      terraform plan
      ```

4. To apply the new configuration, run:

      ```bash
      terraform apply --auto-approve
      ```

## Migrate teams to `aiven_organization_user_group`

Teams have been replaced by groups. Groups are an easier way to
control access to your organization's projects and services for a group of users.
To make the transition to groups smoother, you can migrate your teams to groups.

To migrate from teams to groups:

1. For each team in your organization, make a note of:

   * which users are members of the team
   * which projects the team is assigned to
   * the team's role for each project

2. For each team, create a group with the same name. The following
   sample creates a group using the
   [`aiven_organization_user_group` resource](https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/organization_user_group).

      ```hcl
      resource "aiven_organization_user_group" "admin" {
        organization_id = data.aiven_organization.main.id
        name       = "Admin user group"
        description = "Administrators"
      }
      ```

      -> **Note**
      Users on the Account Owners team automatically become super admin with full access to
      manage the organization. You don't need to create a group for these users or manage
      this team after the migration, and you also can't delete this team.

3. To add the users to the groups, use the
   [`aiven_organization_user_group_member` resource](https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/organization_user_group_member):

      ```hcl
      resource "aiven_organization_user_group_member" "admin_members" {
        group_id         = aiven_organization_user_group.admin.group_id
        organization_id  = data.aiven_organization.main.id
        user_id          = one([for user in data.aiven_organization_user_list.users: user if user.email == "izumi@example.com" ])
      }
      ```

      You can [use the `aiven_organization_user_list` data source](https://docs.aiven.io/docs/platform/reference/get-resource-IDs) to get
      the `user_id`.

4. To add each new group to the same projects that the teams are assigned to, use the
   [`aiven_organization_permission` resource](https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/organization_permission):

      ```hcl
      resource "aiven_organization_permission" "project_admin" {
        organization_id = data.aiven_organization.main.id
        resource_id     = data.aiven_project.example_project.project
        resource_type   = "project"
        permissions {
          permissions = [
            "admin"
          ]
          principal_id   = aiven_organization_user_group.admin.group_id
          principal_type = "user_group"
        }
      }
      ```

5. Preview your changes by running:

      ```bash
      terraform plan
      ```

6. To apply the new configuration, run:

      ```bash
      terraform apply --auto-approve
      ```

7. After confirming all users have the correct access, delete the team resources.

   ~> **Important**
   Deleting the teams in your organization will disable the teams feature. You won't be able to create new teams or access your Account Owners team.

## Update `aiven_redis` resources after Valkey upgrade

After you [upgrade from Aiven for Caching to Aiven for Valkey™](https://aiven.io/docs/products/caching/howto/upgrade-aiven-for-caching-to-valkey), update your
Terraform configuration to use the `aiven_valkey` resource. Aiven for Caching can only be upgraded to Valkey using the Aiven Console or the Aiven API.

The following steps show you how to update your Terraform files using this example file with an Aiven for Caching service:

```hcl
resource "aiven_redis" "caching_service" {
 project      = data.aiven_project.example_project.project
 cloud_name   = "google-europe-west1"
 plan         = "business-4"
 service_name = "example-caching-service"

 redis_user_config {
   redis_timeout = 120
   redis_maxmemory_policy = "allkeys-random"
 }
}

resource "aiven_redis_user" "caching_example_user" {
  service_name = aiven_redis.caching_service.service_name
  project      = data.aiven_project.example_project.project
  username     = "example-user"
  password     = var.caching_user_pw
}
```

1. Replace the `aiven_redis` resources with `aiven_valkey`:

      ```hcl
      resource "aiven_valkey" "caching_service" {
       project      = data.aiven_project.example_project.project
       cloud_name   = "google-europe-west1"
       plan         = "business-4"
       service_name = "example-caching-service"

       valkey_user_config {
         valkey_timeout = 120
         valkey_maxmemory_policy = "allkeys-random"
       }
      }
      ```

2. Replace `aiven_redis_user` resources with `aiven_valkey_user`:

      ```hcl
      resource "aiven_valkey_user" "caching_example_user" {
         service_name = aiven_valkey.caching_service.service_name
         project      = data.aiven_project.example_project.project
         username     = "example-user"
         password     = var.caching_user_pw
      }
      ```

3. To remove Terraform's control of the aiven_redis resources, run:

      ```bash
      terraform state rm aiven_redis.caching_service
      terraform state rm aiven_redis_user.caching_example_user
      ```

4. Add the Valkey resources to Terraform by [importing them](https://registry.terraform.io/providers/aiven/aiven/latest/docs/guides/importing-resources).
   For example:

      ```bash
      terraform import aiven_valkey.caching_service PROJECT/example-caching-service
      terraform import aiven_valkey_user.caching_example_user PROJECT/example-caching-service/example-user
      ```

     Where `PROJECT` is the name of the project.

5. To preview the changes, run:

      ```bash
      terraform plan
      ```

6. To apply the changes, run:

      ```bash
      terraform apply --auto-approve
      ```

7. To confirm the changes, list the resources in the state file by running:

      ```bash
      terraform state list
      ```

## Migrate from M3DB to Thanos Metrics

Migrate your Aiven for M3 databases to [Aiven for Thanos Metrics](https://aiven.io/docs/products/metrics).

1. Create an Aiven for Thanos Metrics service to migrate your Aiven for M3 databases to using the `aiven_thanos` resource:
      ```hcl
      resource "aiven_thanos" "example_thanos" {
       project      = data.aiven_project.example_project.project
       cloud_name   = "google-europe-west1"
       plan         = "business-4"
       service_name = "example-thanos-service"
      }
      ```

2. In the Aiven Console, [migrate your M3DB database to this Thanos service](https://aiven.io/docs/products/metrics/howto/migrate-m3db-thanos).

3. After the migration, remove the `aiven_m3db` and `aiven_m3db_user` resources.

     -> **Note**
      Aiven for Metrics does not have service users. You can grant access to the service using [project roles and permissions](https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/organization_permission).

4. To preview the changes, run:

      ```bash
      terraform plan
      ```

5. To apply the changes, run:

      ```bash
      terraform apply --auto-approve
      ```

## Migrate from `timeouts.default`

The `timeouts.default` field is deprecated and will be removed in a future version. The [Terraform Plugin Framework does not support the `default` timeout field](https://developer.hashicorp.com/terraform/plugin/framework/resources/timeouts).

### What changed

Resources that previously used `timeouts.default` now require specific CRUD timeouts:

**Before (deprecated):**
```hcl
resource "aiven_pg" "example" {
  project      = "my-project"
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "my-postgres"

  timeouts {
    default = "20m"  # This is deprecated
  }
}
```

**After (recommended):**
```hcl
resource "aiven_pg" "example" {
  project      = "my-project"
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "my-postgres"

  timeouts {
    create = "20m"
    read   = "5m"
    update = "20m"
    delete = "20m"
  }
}
```

### Migration steps

1. Replace `timeouts.default` with specific CRUD timeouts in your Terraform configuration
2. Set appropriate timeouts for each operation based on your needs:
    - `create`: Time for resource creation (typically longer)
    - `read`: Time for reading resource state (typically shorter)
    - `update`: Time for resource updates (typically longer)
    - `delete`: Time for resource deletion (typically medium)

-> **Note**
You'll see deprecation warnings during `terraform plan` and `terraform apply` operations until you migrate to specific CRUD timeouts.
