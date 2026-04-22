data "aiven_mysql_database" "example" {
  project       = "my-project"
  service_name  = "my-mysql"
  database_name = "testdb"
}
