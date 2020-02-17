# Project
resource "aiven_project" "foo" {
  project = "project-1"
  account_id = aiven_account.foo.account_id
}