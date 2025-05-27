# Define a random_string resource to generate a suffix
resource "random_string" "suffix" {
  length  = 8
  special = false
  upper   = false
  numeric = true
}

# Create a project within an organization
resource "aiven_organization_project" "example_project" {
  project_id       = "example-project-${random_string.suffix.result}"
  organization_id  = aiven_organization.main.id
  parent_id        = aiven_organization.main.id
  billing_group_id = aiven_billing_group.main.id

  tag {
    key   = "env"
    value = "prod"
  }
}

# Create a project within an organizational unit
resource "aiven_organization_project" "example_project" {
  project_id       = "example-project-in-unit-${random_string.suffix.result}"
  organization_id  = aiven_organization.main.id
  parent_id        = data.aiven_organizational_unit.example_unit.id
  billing_group_id = aiven_billing_group.main.id
}