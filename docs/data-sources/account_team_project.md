---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_account_team_project Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  The Account Team Project data source provides information about the existing Account Team Project.
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

# aiven_account_team_project (Data Source)

The Account Team Project data source provides information about the existing Account Team Project.

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
data "aiven_account_team_project" "account_team_project1" {
  account_id   = aiven_account.<ACCOUNT_RESOURCE>.account_id
  team_id      = aiven_account_team.<TEAM_RESOURCE>.team_id
  project_name = aiven_project.<PROJECT>.project
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `account_id` (String) The unique account id
- `project_name` (String) The name of an already existing project
- `team_id` (String) An account team id

### Read-Only

- `id` (String) The ID of this resource.
- `team_type` (String) The Account team project type. The possible values are `admin`, `developer`, `operator`, `organization:app_users:write`, `organization:audit_logs:read`, `organization:billing:read`, `organization:billing:write`, `organization:domains:write`, `organization:groups:write`, `organization:networking:read`, `organization:networking:write`, `organization:projects:write`, `organization:users:write`, `project:audit_logs:read`, `project:integrations:read`, `project:integrations:write`, `project:networking:read`, `project:networking:write`, `project:permissions:read`, `project:services:read`, `project:services:write`, `read_only`, `role:organization:admin`, `role:services:maintenance`, `role:services:recover`, `service:configuration:write`, `service:data:write`, `service:logs:read`, `service:secrets:read` and `service:users:write`.
