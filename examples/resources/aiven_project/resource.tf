resource "aiven_project" "example_project" {
  project    = "example-project"
  parent_id = aiven_organization.main.id
}
