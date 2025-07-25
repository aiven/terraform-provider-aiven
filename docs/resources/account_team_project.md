---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_account_team_project Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Links an existing project to an existing team. Both the project and team should have the same account_id.
  !> Teams have been deprecated and are being migrated to groups
  On 30 September 2024 the Account Owners team will transition to super admin. Super admin have full access to the organization.
  The Account Owners and super admin are synced, so the removal of the Account Owners team will have no impact on existing permissions.
  From 4 November 2024 you won't be able to create new teams or update existing ones. Existing teams will be migrated to groups after
  this date. On 2 December 2024 all teams will be deleted and the teams feature will be completely removed. View the
  migration guide https://aiven.io/docs/tools/terraform/howto/migrate-from-teams-to-groups for more information on the changes and migrating to groups.
  ~> Important
  You can't delete the Account Owners team. Deleting all other teams in your organization will disable the teams feature.
  You won't be able to create new teams or access your Account Owners team.
---

# aiven_account_team_project (Resource)

Links an existing project to an existing team. Both the project and team should have the same `account_id`.


!> **Teams have been deprecated and are being migrated to groups**
**On 30 September 2024** the Account Owners team will transition to super admin. Super admin have full access to the organization.
The Account Owners and super admin are synced, so the removal of the Account Owners team will have no impact on existing permissions.
**From 4 November 2024** you won't be able to create new teams or update existing ones. Existing teams will be migrated to groups after
this date. **On 2 December 2024** all teams will be deleted and the teams feature will be completely removed. [View the
migration guide](https://aiven.io/docs/tools/terraform/howto/migrate-from-teams-to-groups) for more information on the changes and migrating to groups.

~> **Important**
You can't delete the Account Owners team. **Deleting all other teams in your organization will disable the teams feature.**
You won't be able to create new teams or access your Account Owners team.

## Example Usage

```terraform
resource "aiven_project" "example_project" {
  project    = "project-1"
  account_id = aiven_account_team.ACCOUNT_RESOURCE_NAME.account_id
}

resource "aiven_account_team" "example_team" {
  account_id = aiven_account.ACCOUNT_RESOURCE_NAME.account_id
  name       = "Example team"
}

resource "aiven_account_team_project" "main" {
  account_id   = aiven_account.ACCOUNT_RESOURCE_NAME.account_id
  team_id      = aiven_account_team.example_team.team_id
  project_name = aiven_project.example_project.project
  team_type    = "admin"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `account_id` (String) The unique account id
- `team_id` (String) An account team id

### Optional

- `project_name` (String) The name of an already existing project
- `team_type` (String) The Account team project type. The possible values are `admin`, `developer`, `operator`, `organization:app_users:write`, `organization:audit_logs:read`, `organization:billing:read`, `organization:billing:write`, `organization:domains:write`, `organization:groups:write`, `organization:networking:read`, `organization:networking:write`, `organization:projects:write`, `organization:users:write`, `project:audit_logs:read`, `project:integrations:read`, `project:integrations:write`, `project:networking:read`, `project:networking:write`, `project:permissions:read`, `project:services:read`, `project:services:write`, `read_only`, `role:organization:admin`, `role:services:maintenance`, `role:services:recover`, `service:configuration:write`, `service:data:write`, `service:logs:read`, `service:secrets:read` and `service:users:write`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of this resource.

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
terraform import aiven_account_team_project.account_team_project1 account_id/team_id/project_name
```
