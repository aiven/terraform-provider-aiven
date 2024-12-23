resource "aiven_project" "example_project" {
  project    = "project-1"
  account_id = aiven_account_team.ACCOUNT_RESOURCE_NAME.account_id
}

resource "aiven_account_team" "example_team" {
  account_id = aiven_account.ACCOUNT_RESOURCE_NAME.account_id
  name       = "Example team"
}

resource "aiven_account_team_project" "main" {
  account_id   = aiven_account.ACCOUNT_RESOURCE_NAME.account_id
  team_id      = aiven_account_team.example_team.team_id
  project_name = aiven_project.example_project.project
  team_type    = "admin"
}
