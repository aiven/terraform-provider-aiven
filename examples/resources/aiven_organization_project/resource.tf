resource "aiven_organization_project" "example" {
  organization_id  = "org1234abcd"
  project_id       = "project-1" // Force new
  billing_group_id = "721bf796-1d89-402d-9195-425a23c4efdc"
  parent_id        = "a3fd7a594e01"

  // OPTIONAL FIELDS
  base_port = 42
  tag {
    key   = "foo"
    value = "foo"
  }
  technical_emails = ["foo@example.com"]

  /* COMPUTED FIELDS
  ca_cert = "foo"
  */
}
