resource "aiven_organization_application_user" "tf_user" {
  organization_id = aiven_organization.main.id
  name            = "app-terraform"
}

resource "aiven_organization_application_user_token" "example" {
  organization_id = aiven_organization.main.id
  user_id         = aiven_organization_application_user.tf_user.user_id
  description     = "Token for TF access to Aiven."
}
