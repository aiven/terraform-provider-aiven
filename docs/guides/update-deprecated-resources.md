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
   To list all resources in the state file, run: `terraform state list`

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
This example shows you how to migrate to the new permission resource.

The following file has a user assigned to a project with the read only role and a group assigned to the same project with the operator role.

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

   # Grant the read_only role to the user and the operator role to the group
   resource "aiven_organization_permission" "example_permissions" {
     organization_id = data.aiven_organization.main.id
     resource_id     = data.aiven_project.example_project.id
     resource_type   = "project"
     permissions {
       permissions = [
         "read_only"
       ]
       principal_id   = "u123a456b7890c"
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

   Where `principal_id` is the user or group ID. You can [get the IDs from the Aiven Console](https://docs.aiven.io/docs/platform/reference/get-resource-IDs).

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

Teams have been deprecated and are being migrated to groups. Groups are an easier way to 
control access to your organization's projects and services for a group of users.

* **On 30 September 2024 the Account Owners team will be removed.**

  The Account Owners and super admin are synced, so the removal of the
  Account Owners team will have no impact on existing permissions.
  Super admin have full access to organizations.

* **From 4 November 2024 you won't be able to create new teams or update existing ones.**

  To simplify the move, Aiven will also begin migrating your existing teams to groups.

* **On 2 December 2024 all teams will be migrated to groups and deleted.**

  To make the transition to groups smoother, you can
  migrate your teams before this date.

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
     user_id          = "u123a456b7890c"
   }
   ```

   You can [get IDs from the Aiven Console](https://docs.aiven.io/docs/platform/reference/get-resource-IDs).

4. To add each new group to the same projects that the teams are assigned to, use the
   [`aiven_organization_permission` resource](https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/organization_permission):

   ```hcl
   resource "aiven_organization_permission" "project_admin" {
     organization_id = data.aiven_organization.main.id
     resource_id     = data.aiven_project.example_project.id
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


## Update teams resources

Teams have been replaced by groups. **On 2 December 2024 all existing teams will be migrated to groups and deleted.**

If you didn't remove your teams, then the automatic migration will create
groups with the same names, users, and assigned projects. After the automatic migration
from teams to groups you will need to update your Terraform files with the groups resources.

The following steps show you how to update your Terraform files to
replace your team resources with the new groups. The example uses
this file with a team that has one member and one project:

```hcl
terraform {
  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = ">=1.1.0, <1.1.0"
    }
  }
}

provider "aiven" {
  api_token = var.aiven_token
}

# Your account
data "aiven_account" "main" {
  name = "Example Account"
}

# Your project
data "aiven_project" "example_project" {
  project = "example-project"
}

# Team
resource "aiven_account_team" "example_team" {
  account_id = data.aiven_account.main.account_id
  name       = "Example team"
}

# Team member
resource "aiven_account_team_member" "example_team_member" {
  account_id = data.aiven_account.main.account_id
  team_id    = aiven_account_team.example_team.team_id
  user_email = "amal@example.com"
}

# Team added to the project
resource "aiven_account_team_project" "main" {
  account_id   = data.aiven_account.main.account_id
  team_id      = aiven_account_team.example_team.team_id
  project_name = data.aiven_project.example_project.project
  team_type    = "admin"
}
```

1. Replace the `aiven_account_team` resources with
   `aiven_organization_user_group`:

   ```hcl
   # The new group created by Aiven from a team of the same name.
   resource "aiven_organization_user_group" "example_group" {
    name            = "Example group"
    description     = ""
    organization_id = "org1a23f456789"
    }
   ```

   You can [get the IDs](https://docs.aiven.io/docs/platform/reference/get-resource-IDs) in the Aiven Console.

2. Replace the `aiven_account_team_member` resources with
   `aiven_organization_user_group_member`:

   ```hcl
   resource "aiven_organization_user_group_member" "project_admin" {
       group_id        = aiven_organization_user_group.example_group.group_id
       organization_id = "org1a23f456789"
       user_id         = "u123a456b7890c"
    }
   ```

3. Replace the `aiven_account_team_project` resources with
    `aiven_organization_permission`:

      ```hcl
      resource "aiven_organization_permission" "main" {
        organization_id = data.aiven_organization.main.id
        resource_id     = data.aiven_project.example_project.id
        resource_type   = "project"
        permissions {
          permissions = [
            "admin"
          ]
          principal_id   = aiven_organization_user_group.example_group.group_id
          principal_type = "user_group"
        }
      }
      ```

4. To remove Terraform's control of the team resources in this list run:

   ```bash
   terraform state rm aiven_account_team.example_team
   terraform state rm aiven_account_team_member.example_team_member
   terraform state rm aiven_account_team_project.main
   ```

5. Add the group resources to Terraform by [importing them](https://registry.terraform.io/providers/aiven/aiven/latest/docs/guides/importing-resources).
    * For groups, run:

      ```bash
      terraform import aiven_organization_user_group.example_group ORGANIZATION_ID/USER_GROUP_ID
      ```

    * For group members, run:
      ```bash
      terraform import aiven_organization_user_group_member.project_admin ORGANIZATION_ID/USER_GROUP_ID/USER_ID
      ```

    * For projects assigned to the groups:

      ```bash
      terraform import aiven_organization_group_project.main PROJECT/USER_GROUP_ID
      ```

    Where:
    * `ORGANIZATION_ID` is the ID of the organization the group is in.
    * `USER_GROUP_ID` is the ID of the user group in the format `ug123a456b7890c`.
    * `USER_ID` is the ID of the user in the format `u123a456b7890c`.
    * `PROJECT` is the name of the project.

    You can [get IDs in the Aiven Console](https://docs.aiven.io/docs/platform/reference/get-resource-IDs).

6. To preview the changes, run:

   ```bash
   terraform plan
   ```

   -> The user group resources will be replaced. This is expected behavior.

7. To apply the changes, run:

   ```bash
   terraform apply --auto-approve
   ```

8. To confirm the changes, list the resources in the state file by running:

   ```bash
   terraform state list
   ```

## Update `aiven_redis` resources after Valkey upgrade

After you [upgrade from Aiven for Caching to Aiven for Valkey™](https://aiven.io/docs/products/caching/howto/upgrade-aiven-for-caching-to-valkey), update your
Terraform configraution to use the `aiven_valkey` resource. Aiven for Caching can only be upgraded to Valkey using the Aiven Console or the Aiven API.

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
