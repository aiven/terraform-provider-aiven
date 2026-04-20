data "aiven_organization_user_list" "example" {
  // REQUIRED EXACTLY ONE
  id      = "org1a23f456789"
  // name = "foo"

  /* COMPUTED FIELDS
  users {
    user_id            = "foo"
    is_super_admin     = true
    join_time          = "2021-01-01T00:00:00Z"
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
