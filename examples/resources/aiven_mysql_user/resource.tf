resource "aiven_mysql_user" "foo" {
  service_name = aiven_mysql.bar.service_name
  project      = "my-project"
  username     = "user-1"
  password     = "Test$1234"
}