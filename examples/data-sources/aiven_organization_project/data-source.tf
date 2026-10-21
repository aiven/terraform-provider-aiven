data "aiven_organization_project" "example" {
  organization_id = "org1234abcd"
  project_id      = "project-1"

  /* COMPUTED FIELDS
  billing_group_id = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  parent_id        = "a3fd7a594e01"
  base_port        = 10000
  ca_cert          = "foo"
  tag {
    key   = "foo"
    value = "foo"
  }
  technical_emails = ["foo@example.com"]
  */
}
