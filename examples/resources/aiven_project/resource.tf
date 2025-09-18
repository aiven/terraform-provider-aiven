# Create a project within an organization
resource "aiven_project" "example_project" {
  project   = "example-project"
  parent_id = aiven_organization.main.id
}

# Create a project within an organizational unit
resource "aiven_project" "example_project" {
  project_id       = "example-project"
  parent_id        = aiven_organizational_unit.example_unit.id
}
