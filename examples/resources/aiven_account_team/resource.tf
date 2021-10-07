resource "aiven_account_team" "account_team1" {
    account_id = aiven_account.<ACCOUNT_RESOURCE>.account_id
    name = "<ACCOUNT_TEAM_NAME>"
}
