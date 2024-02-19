resource "aiven_account_team_member" "main" {
  account_id = aiven_account.ACCOUNT_RESOURCE_NAME.account_id
  team_id    = aiven_account_team.TEAM_RESOURCE_NAME.team_id
  user_email = "user+1@example.com"
}
