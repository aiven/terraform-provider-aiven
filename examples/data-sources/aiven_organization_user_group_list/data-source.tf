data "aiven_organization_user_group_list" "example" {
  organization_id = "org1a23f456789"

  /* COMPUTED FIELDS
  user_groups {
    user_group_id   = "foo"
    create_time     = "2021-01-01T00:00:00Z"
    description     = "example description"
    managed_by_scim = true
    member_count    = 42
    update_time     = "2021-01-01T00:00:00Z"
    user_group_name = "foo"
  }
  */
}
