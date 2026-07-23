resource "aiven_organization_user_group" "example" {
  organization_id = "org1a23f456789" // Force new
  description     = "The group of admins for the organization"
  name            = "Admin Users"

  /* COMPUTED FIELDS
  group_id        = "foo"
  create_time     = "2021-01-01T00:00:00Z"
  managed_by_scim = true
  update_time     = "2021-01-01T00:00:00Z"
  */
}
