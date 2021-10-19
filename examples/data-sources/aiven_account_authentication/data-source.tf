data "aiven_account_authentication" "foo" {
    account_id = aiven_account.<ACCOUNT_RESOURCE>.account_id
    name = "auth-1"
}
