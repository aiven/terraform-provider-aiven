
resource "aiven_organization_user_group" "example" {
  description = "Example group of users."
  organization_id = aiven_organization.main.id
  name = "Example group"
}

resource "aiven_organization_user_group_project" "example" {
  group_id = aiven_organization_user_group.example.group_id
  project = aiven_project.example.project
  role = "admin"
}