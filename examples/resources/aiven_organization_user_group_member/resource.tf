resource "aiven_organization_user_group_member" "example" {
  organization_id = "org1a23f456789" // Force new
  group_id        = "foo" // Force new
  user_id         = "foo" // Force new

  /* COMPUTED FIELDS
  last_activity_time = "2021-01-01T00:00:00Z"
  */
}
