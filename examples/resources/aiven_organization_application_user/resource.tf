resource "aiven_organization_application_user" "tf_user" {
  organization_id = aiven_organization.main.id
  name            = "app-terraform"
}
