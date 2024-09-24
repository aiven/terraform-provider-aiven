# Grant permission to a user
resource "aiven_organization_permission" "operator" {
  organization_id = data.aiven_organization.main.id
  resource_id     = data.aiven_project.example_project.id
  resource_type   = "project"
  permissions {
    permissions = [
      "operator"
    ]
    principal_id   = "u123a456b7890c"
    principal_type = "user"
  }
}

# Grant permission to a group
resource "aiven_organization_permission" "developers" {
  organization_id = data.aiven_organization.main.id
  resource_id     = data.aiven_project.example_project.id
  resource_type   = "project"
  permissions {
    permissions = [
      "developer"
    ]
    principal_id   = data.aiven_organization_user_group.example_group.group_id
    principal_type = "user_group"
  }
}