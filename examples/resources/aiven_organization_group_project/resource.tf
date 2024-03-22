resource "aiven_project" "example_project" {
  project    = "Example project"
  parent_id = aiven_organization.main.id
}

resource "aiven_organization_user_group" "example" {
  description = "Example group of users."
  organization_id = aiven_organization.main.id
  name = "Example group"
}

resource "aiven_organization_user_group_member" "project_admin" {
  group_id = aiven_organization_user_group.example.group_id
  organization_id = aiven_organization.main.id
  user_id = "u123a456b7890c" 
}

resource "aiven_organization_group_project" "example" {
  group_id = aiven_organization_user_group.example.group_id
  project = aiven_project.example_project.project
  role = "admin"
}