data "aiven_organization_user_group" "example" {
  organization_id = "org1a23f456789"

  // LOOKUP — provide `group_id` or `name`
  group_id = "foo"
  // name  = "Admin Users"

  /* COMPUTED FIELDS
  create_time     = "2021-01-01T00:00:00Z"
  description     = "The group of admins for the organization"
  managed_by_scim = true
  update_time     = "2021-01-01T00:00:00Z"
  */
}
