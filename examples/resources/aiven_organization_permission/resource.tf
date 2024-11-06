resource "aiven_organization_permission" "example_permissions" {
  organization_id = data.aiven_organization.main.id
  resource_id     = data.aiven_project.example_project.id
  resource_type   = "project"
  permissions {
    # Grant the operator role and permission to read service logs to a user
    permissions = [
      "operator",
      "service:logs:read"
    ]
    principal_id   = "u123a456b7890c"
    principal_type = "user"
  }
  # Grant write project integrations and read project networking permissions, and the developer role to a group
  permissions {
    permissions = [
      "project:integrations:write",
      "project:networking:read",
      "developer"
    ]
    principal_id   = data.aiven_organization_user_group.example_group.group_id
    principal_type = "user_group"
  }
}
