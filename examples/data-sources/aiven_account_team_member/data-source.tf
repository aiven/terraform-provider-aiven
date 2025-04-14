data "aiven_account_team_member" "foo" {
  account_id = aiven_account.ACCOUNT_RESOURCE.account_id
  team_id    = aiven_account_team.TEAM_RESOURCE.team_id
  user_email = "user+1@example.com"
}
