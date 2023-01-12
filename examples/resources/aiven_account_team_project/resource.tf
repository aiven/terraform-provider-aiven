resource "aiven_project" "<PROJECT>" {
  project    = "project-1"
  account_id = aiven_account_team.<ACCOUNT_RESOURCE>.account_id
}

resource "aiven_account_team_project" "account_team_project1" {
  account_id   = aiven_account.<ACCOUNT_RESOURCE>.account_id
  team_id      = aiven_account_team.<TEAM_RESOURCE>.team_id
  project_name = aiven_project.<PROJECT>.project
  team_type    = "admin"
}
