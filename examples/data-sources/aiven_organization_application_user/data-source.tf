data "aiven_organization_application_user" "tf_user" {
  organization_id = aiven_organization.main.id
  user_id         = "u123a456b7890c"
}

# Alternatively, you can use the aiven_organization_user_list data source
# to get the user_id of an application user by its name
data "aiven_organization_user_list" "users" {
  name = "Example organization"
}

data "aiven_organization_application_user" "app_user" {
  organization_id = data.aiven_organization.main.id
  user_id         = one([for user in data.aiven_organization_user_list.users.users : user.user_id if user.user_info[0].real_name == "app-user"])
}
