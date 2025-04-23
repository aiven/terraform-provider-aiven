# Grant access to a specific project
resource "aiven_organization_permission" "example_project_permissions" {
  organization_id = data.aiven_organization.main.id
  resource_id     = data.aiven_project.example_project.project
  resource_type   = "project"
  permissions {
    # Grant a user the operator role and
    # permission to read service logs
    permissions = [
      "operator",
      "service:logs:read"
    ]
    principal_id   = "u123a456b7890c"
    principal_type = "user"
  }
  # Grant a group the write project integrations
  # permission and the developer role
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

  # Grant a user permission to manage application
  # users and view all project audit logs
  permissions {
    permissions = [
      "organization:app_users:write",
      "project:audit_logs:read"
    ]
    principal_id   = "u123a456b7890c"
    principal_type = "user"
  }

  # Grant a group permission to manage users,
  # groups, domains, and identity providers
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
