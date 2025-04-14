resource "aiven_organization_project" "foo" {
  project_id = "example-project"

  organization_id  = aiven_organization.foo.id
  billing_group_id = aiven_billing_group.foo.id

  tag {
    key   = "key_1"
    value = "value_1"
  }
}
