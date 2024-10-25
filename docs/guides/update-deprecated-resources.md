---
page_title: "Update deprecated resources"
---

# Update deprecated resources

Migrate from resources that have been deprecated or renamed.

## Migrate renamed resources

Migrate renamed resources without destroying your existing resources.

1.  Optional: Backup your Terraform state file, `terraform.tfstate`, to use in the case
    of a rollback.

2.  Replace references to the deprecated resource with the new name. In
    the following example, the `aiven_database` resource was replaced with
    `aiven_pg_database`:

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

3.  Preview removing the old resource from Terraform's control by running:

    ```bash
    terraform state rm -dry-run RESOURCE.RESOURCE_NAME
    ```

    For example, to preview removing the `aiven_database` resource named `mydatabase`, run:

    ```bash
    terraform state rm -dry-run aiven_database.mydatabase
    ```

4.  To remove the resource, run the command without the `-dry-run` flag:

    ```bash
    terraform state rm RESOURCE.RESOURCE_NAME
    ```

5.  Add the resource back to Terraform with the new name by [importing it](https://registry.terraform.io/providers/aiven/aiven/latest/docs/guides/importing-resources).

    For example, to add the `aiven_pg_database` resource, run:

    ```bash
    terraform import aiven_pg_database.mydatabase myproject/mypgservice/example-database
    ```

6.  Preview your changes by running:

    ```bash
    terraform plan
    ```

7.  To apply the new configuration, run:

    ```bash
    terraform apply --auto-approve
    ```

## Migrate deprecated resources

1.  Optional: Backup your Terraform state file, `terraform.tfstate`, to use in the case
    of a rollback.

2.  Replace all instances of the deprecated resource with the new resource in your Terraform files.

    -> **Tip**
    To list all resources in the state file, run: `terraform state list`

3.  Preview removing the deprecated resource from Terraform's control by running:

    ```bash
    terraform state rm -dry-run RESOURCE.RESOURCE_NAME
    ```

4.  To remove the resource, run the command without the `-dry-run` flag:

    ```bash
    terraform state rm RESOURCE.RESOURCE_NAME
    ```

5.  To preview your changes, run:

    ```bash
    terraform plan
    ```

6.  To add the new resource, apply the new configuration by running:

    ```bash
    terraform apply --auto-approve
    ```

## Examples

### Migrate to `aiven_organization_permission`

The `aiven_project_user` and `aiven_organization_group_project` resources have been replaced by
the `aiven_organization_permission` resource. The following example shows you how to migrate to the new
permissions resource.

The following file has a user assigned to a project with the read_only role and a group assigned to the same project with the operator role.

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

# Assign a user to a project
resource "aiven_project_user" "example_project_user" {
  project     = data.aiven_project.example_project.project
  email       = "dana@example.com"
  member_type = "read_only"
}

# Assign a group to project with the operator role
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

   # New permissions resource granting the read_only role to the user
   resource "aiven_organization_permission" "operator" {
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
   }

   # New permissions resource granting the operator role to the group
   resource "aiven_organization_permission" "example" {
     organization_id = data.aiven_organization.main.id
     resource_id     = data.aiven_project.example_project.id
     resource_type   = "project"
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
    terraform state rm RESOURCE.RESOURCE_NAME
    ```

    For example:

    ```bash
    terraform state rm aiven_organization_group_project.example
    terraform state rm aiven_project_user.example_project_user
    ```

3. Preview your changes by running:

    ```bash
    terraform plan
    ```

4.  To apply the new configuration, run:

    ```bash
    terraform apply --auto-approve
    ```

### Migrate to `aiven_organization_user_group`

Teams have been deprecated and are being migrated to groups. Groups are an easier way to control access to your organization's projects and services for a group of users.

* **On September 30, 2024 the Account Owners team will be removed.**

  The Account Owners and super admin are synced, so the removal of the
  Account Owners team will have no impact on existing permissions.
  [Super admin](/docs/platform/concepts/orgs-units-projects#users-and-roles)
  have full access to organizations.

* **From November 4, 2024 you won't be able to create new teams or update existing ones.**

  To simplify the move, Aiven will also begin migrating your existing teams to groups.

* **On December 2, 2024 all teams will be migrated to groups and deleted.**

  To make the transition to groups smoother, you can
  migrate your teams before this date.

~> **Important**
You can't delete the Account Owners team. **Deleting all other teams in your organization
will disable the teams feature.** You won't be able to create new teams or access your
Account Owners team.

To migrate from teams to groups:

1.  For each team, make a note of:

    * which users are members of the team
    * which projects the team is assigned to
    * the team's role for each project

2.  For each team in your organization, create a group with the same name. The following
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
    this team after the migration.

3.  To add the users to the groups, use the
    [`aiven_organization_user_group_member` resource](https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/organization_user_group_member):

    ```hcl
    resource "aiven_organization_user_group_member" "admin_members" {
      group_id      = aiven_organization_user_group.admin.group_id
      organization_id = data.aiven_organization.main.id
      user_id = "u123a456b7890c"
    }
    ```

4.  To add each new group to the same projects that the teams are assigned to, use the
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

5.  Preview your changes by running:

    ```bash
    terraform plan
    ```

6.  To apply the new configuration, run:

    ```bash
    terraform apply --auto-approve
    ```

7.  After confirming all users have the correct access, delete the team resources.
