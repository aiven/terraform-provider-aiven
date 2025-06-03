data "aiven_organization_user_group" "example" {
  name            = "Example group"
  organization_id = aiven_organization.main.id
}
