data "aiven_organization_application_user" "tf_user" {
  organization_id = aiven_organization.main.id
  user_id = "u123a456b7890c"
}
