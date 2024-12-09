# Project-level permissions
# Grant access to a specific project
resource "aiven_organization_permission" "example_project_permissions" {
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
  # Grant write project integrations, and the developer role to a group
  permissions {
    permissions = [
      "project:integrations:write",
      "developer"
    ]
    principal_id   = data.aiven_organization_user_group.example_group.group_id
    principal_type = "user_group"
  }
}

# Organization-level permissions
resource "aiven_organization_permission" "example_org_permissions" {
  organization_id = data.aiven_organization.main.id
  resource_id     = data.aiven_organization.main.id
  resource_type   = "organization"

  # Grant access to manage application users and 
  # view all project audit logs to a user
  permissions {
    permissions = [
      "organization:app_users:write",
      "project:audit_logs:read"
    ]
    principal_id   = "u123a456b7890c" 
    principal_type = "user"
  }

  # Grant access to users, groups, domains, and
  # identity providers to a group
  permissions {
    permissions = [
      "organization:users:write",
      "organization:groups:write",
      "organization:domains:write",
      "organization:idps:write"
    ]
    principal_id   = aiven_organization_user_group.example_group.group_id
    principal_type = "user_group"
  }
}
