# Account
resource "aiven_account" "foo" {
  name = "account1"
}

# Account team
resource "aiven_account_team" "foo" {
  account_id = aiven_account.foo.account_id
  name = "account_team1"
}

# Account team project
resource "aiven_account_team_project" "foo" {
  account_id = aiven_account.foo.account_id
  team_id = aiven_account_team.foo.team_id
  project_name = aiven_project.foo.project
  team_type = "admin"
}

# Account team member
resource "aiven_account_team_member" "foo" {
  account_id = aiven_account.foo.account_id
  team_id = aiven_account_team.foo.team_id
  user_email = "user+1@example.com"
}

data "aiven_account_team" "team" {
  name = aiven_account_team.foo.name
  account_id = aiven_account_team.foo.account_id
}

# Account authentication
resource "aiven_account_authentication" "foo" {
  account_id = aiven_account.foo.account_id
  name = "auth-1"
  type = "saml"
  enabled = false
}