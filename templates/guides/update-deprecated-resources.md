---
page_title: "Migrate resources"
---

# Migrate resources

You can migrate your existing resources to new resources by updating your Terraform configuration and state file. Resources need to be migrated when:

* A resource is deprecated in favor of a replacement. Deprecation notices are shown when planning and applying changes, in the documentation, and in the changelog.
* A resource is changed in another interface like the Aiven Console or API. This can happen when Aiven makes changes like migrating customers from an
  old feature to a new one. It can also happen in cases where you can only make a change, like an upgrade, in one of these other interfaces.
  In these cases, you need to update your Terraform configuration to match the actual state.
* A resource field is deprecated or changed.

Details about changes to resources are available in the [changelog](https://github.com/aiven/terraform-provider-aiven/blob/main/CHANGELOG.md).
Aiven also sends email notifications for situations like automatic migrations to new resources.

## Migrate deprecated resources

To replace resources that are deprecated, add the new resource using the [`import` block](https://developer.hashicorp.com/terraform/language/import).
Import blocks are the recommended approach because you can:

* Version and review imports in pull requests like other Terraform changes
* Generate many `import` blocks from a script instead of manually running import commands
* Preview all pending imports in a single `terraform plan` before applying

1. Back up your Terraform state file, `terraform.tfstate`, so you can restore the previous state if needed.

2. Add the import block in your configuration:

      ```hcl
      import {
        to = TYPE.LABEL
        id = "RESOURCE_ID"
      }
      ```

3. Add the new resource to your configuration with the required fields. For example:

      ```hcl
      resource "aiven_pg_database" "mydatabase" {
          project       = "myproject"
          service_name  = "mypgservice"
          database_name = "example-database"
      }
      ```

      -> **Tip**
      To list all resources in the state file, run: `terraform state list`.

4. To preview the import, run `terraform plan`.
5. To apply the import, run `terraform apply`.
6. To remove the deprecated resource from Terraform's control, run:

      ```bash
      terraform state rm $(terraform state list | grep '^TYPE\.')
      ```

   You can use the `-dry-run` flag to preview the changes before removing the deprecated resources.

7. Remove the deprecated resource blocks from your configuration.
8. To confirm that the resources were migrated and that there is no drift between the configuration and the state, run `terraform plan` again.

## Migrate to `aiven_organization_billing_group`

The `aiven_billing_group` resource has been replaced by
[the `aiven_organization_billing_group` resource](https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/organization_billing_group).

This example shows you how to migrate a billing group to the new resource. The following example has a billing group with billing contact emails, billing emails, VAT ID, and custom invoice text.

```hcl
resource "aiven_billing_group" "example" {
  parent_id = "org1a23f456789"
  name      = "Default billing group for the organization"

  billing_contact_emails = ["jane@example.com"]
  billing_emails         = ["billing@example.com"]
  billing_extra_text     = "Purchase order: PO100018"
  card_id                = "pm4b1ff1ceeaa"
  vat_id                 = "FI12345678"
}
```

### Prerequisites

Addresses are now managed as a separate resource, `aiven_organization_address`. Each address can be assigned as the billing and shipping
address for a billing group. You can get the IDs of your organization's addresses in
[the Aiven Console](https://aiven.io/docs/platform/reference/get-resource-IDs).

### Migrate your billing groups

1. To add the new billing group to Terraform, declare the import in your configuration using an
  [`import` block](https://developer.hashicorp.com/terraform/language/import).
  The ID is in the format `ORG_ID/BILLING_GROUP_ID`. For example:

      ```hcl
      import {
        to = aiven_organization_billing_group.example
        id = "org1a23f456789/00ab1234-5678-9cd0-1ef2-345678g9012a"
      }
      ```

1. For each billing group, add an `aiven_organization_billing_group` resource to your configuration. For example:

      ```hcl
      resource "aiven_organization_billing_group" "example" {
        organization_id     = "org1a23f456789"
        billing_group_name  = "Default billing group for the organization"
        billing_address_id  = "addr4b1ff1ceeaa"
        shipping_address_id = "addr4b1ff1ceeaa"
        billing_contact_emails {
          email = "jane@example.com"
        }
        billing_emails {
          email = "billing@example.com"
        }
        payment_method {
          payment_method_id   = "pm4b1ff1ceeaa"
          payment_method_type = "credit_card"
        }
        custom_invoice_text = "Purchase order: PO100018"
        vat_id              = "FI12345678"
      }
      ```

     -> **Tip**
     To list payment method IDs and types for an organization, use the
     [`aiven_organization_payment_method_list` data source](https://registry.terraform.io/providers/aiven/aiven/latest/docs/data-sources/organization_payment_method_list).

2. To preview the import, run `terraform plan`.
3. To apply the import, run `terraform apply`.
4. To remove the deprecated resource from Terraform's control, run:

      ```bash
      terraform state rm $(terraform state list | grep '^aiven_billing_group\.')
      ```

   You can use the `-dry-run` flag to preview the changes before removing the deprecated resources.

5. Remove the deprecated `aiven_billing_group` resources from your configuration.
6. To confirm that the billing groups were migrated and that there is no drift between the configuration and the state, run `terraform plan` again.

-> **Tip**
 Import blocks let you version and review imports in pull requests. Alternatively,
 you can use the CLI to run the import command outside your configuration:
`terraform import aiven_organization_billing_group.example ORG_ID/BILLING_GROUP_ID`

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
