data "aiven_mysql_database" "example_database" {
  project       = aiven_project.example_project.project
  service_name  = aiven_mysql.example_mysql.service_name
  database_name = "example-database"
}
