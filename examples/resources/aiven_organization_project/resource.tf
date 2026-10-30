resource "aiven_organization_project" "example" {
  organization_id  = "org1234abcd"
  project_id       = "my-project" // Force new
  billing_group_id = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  parent_id        = "a3fd7a594e01"

  // OPTIONAL FIELDS
  base_port = 10000
  tag {
    key   = "foo"
    value = "foo"
  }
  technical_emails = ["foo@example.com"]

  /* COMPUTED FIELDS
  ca_cert = "foo"
  */
}
