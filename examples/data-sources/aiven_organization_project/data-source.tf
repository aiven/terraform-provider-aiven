data "aiven_organization_project" "example" {
  project_id      = aiven_organization_project.foo.project_id
  organization_id = aiven_organization_project.foo.organization_id
}