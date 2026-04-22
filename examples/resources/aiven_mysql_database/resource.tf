resource "aiven_mysql_database" "example" {
  project       = "my-project" // Force new
  service_name  = "my-mysql" // Force new
  database_name = "testdb" // Force new
}
