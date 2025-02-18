# Create a project within an organization
resource "aiven_organization_project" "example_project" {
  project_id       = "example-project"
  organization_id  = aiven_organization.main.id
  billing_group_id = aiven_billing_group.main.id

  tag {
    key   = "env"
    value = "prod"
  }
}

# Create a project within an organizational unit
resource "aiven_organization_project" "example_project" {
  project_id       = "example-project"
  organization_id  = aiven_organization.main.id
  parent_id        = data.organizational_unit.example_unit.id
  billing_group_id = aiven_billing_group.main.id
}