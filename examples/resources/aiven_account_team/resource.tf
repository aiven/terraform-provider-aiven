resource "aiven_account_team" "example_team" {
  account_id = aiven_account.ACCOUNT_RESOURCE_NAME.account_id
  name       = "Example team"
}
