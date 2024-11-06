# Grant the operator role and
# the permission to read service logs to a user
resource "aiven_organization_permission" "operator" {
  organization_id = data.aiven_organization.main.id
  resource_id     = data.aiven_project.example_project.id
  resource_type   = "project"
  permissions {
    permissions = [
      "operator",
      "service:logs:read"
    ]
    principal_id   = "u123a456b7890c"
    principal_type = "user"
  }
}

# Grant the write project integrations permission, read project
# networking permission, and developer role to a group
resource "aiven_organization_permission" "developers" {
  organization_id = data.aiven_organization.main.id
  resource_id     = data.aiven_project.example_project.id
  resource_type   = "project"
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