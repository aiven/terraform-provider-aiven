resource "aiven_organization_permission" "example" {
  organization_id = "org1a23f456789" // Force new
  resource_type   = "organization" // Force new
  resource_id     = "foo" // Force new
  permissions {
    principal_id   = "u12345"
    permissions    = ["read_only"]
    principal_type = "user"

    /* COMPUTED FIELDS
    create_time = "2021-01-01T00:00:00Z"
    update_time = "2021-01-01T00:00:00Z"
    */
  }
}
