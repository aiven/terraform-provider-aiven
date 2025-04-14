resource "aiven_organization_user_group" "example" {
  description     = "Example group of users."
  organization_id = aiven_organization.main.id
  name            = "Example group"
}
