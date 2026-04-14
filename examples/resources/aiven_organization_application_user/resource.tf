resource "aiven_organization_application_user" "example" {
  organization_id = "org1a23f456789" // Force new
  name            = "devops app user"

  /* COMPUTED FIELDS
  user_id     = "foo"
  create_time = "2021-01-01T00:00:00Z"
  email       = "foo@example.com"
  */
}
