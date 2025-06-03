data "aiven_external_identity" "external_id_mapping" {
  external_service_name = "github"
  external_user_id      = "sasha-aiven"
  internal_user_id      = "u123a456b7890c"
  organization_id       = data.aiven_organization.main.id
}

# Alternatively, use the aiven_organization_user_list data source
# to get the user_id of a user by their email address
data "aiven_organization_user_list" "users" {
  name = "Example organization"
}

data "aiven_external_identity" "external_id_mapping" {
  external_service_name = "github"
  external_user_id      = "sasha-aiven"
  internal_user_id      = one([for user in data.aiven_organization_user_list.users.users : user.user_id if user.user_info[0].user_email == "EMAIL_ADDRESS"])
  organization_id       = data.aiven_organization.main.id
}
