resource "aiven_organization_user_group" "example" {
  description     = "Example group of users."
  organization_id = aiven_organization.main.id
  name            = "Example group"
}

# Use the aiven_organization_user_list data source
# to get the user_id of a user by their email address
data "aiven_organization_user_list" "users" {
  name = "Example organization"
}

resource "aiven_organization_user_group_member" "project_admin" {
  group_id        = aiven_organization_user_group.example.group_id
  organization_id = aiven_organization.main.id
  user_id         = one([for user in data.aiven_organization_user_list.users.users : user.user_id if user.user_info[0].user_email == "izumi@example.com"])
}