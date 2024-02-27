resource "aiven_project" "example_project" {
  project    = "Example project"
  parent_id = aiven_organization.main.id
}
