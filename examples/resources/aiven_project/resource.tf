resource "aiven_project" "myproject" {
    project = "<PROJECT_NAME>"
    card_id = "<FULL_CARD_ID/LAST4_DIGITS>"
    account_id = aiven_account_team.<ACCOUNT_RESOURCE>.account_id
}
