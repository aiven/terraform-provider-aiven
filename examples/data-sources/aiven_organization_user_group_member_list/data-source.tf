data "aiven_organization_user_group_member_list" "example" {
  organization_id = "org1a23f456789"
  user_group_id   = "foo"

  /* COMPUTED FIELDS
  members {
    user_id            = "foo"
    last_activity_time = "2021-01-01T00:00:00Z"
    user_info {
      managing_organization_id = "org1a23f456789"
      city                     = "foo"
      country                  = "foo"
      create_time              = "2021-01-01T00:00:00Z"
      department               = "foo"
      is_application_user      = true
      job_title                = "foo"
      managed_by_scim          = true
      real_name                = "foo"
      state                    = "foo"
      user_email               = "foo@example.com"
    }
  }
  */
}
