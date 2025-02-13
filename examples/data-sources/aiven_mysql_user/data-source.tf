data "aiven_mysql_user" "example_mysql_user" {
  service_name = aiven_mysql.example_mysql.service_name
  project      = aiven_project.example_project.project
  username     = "example-mysql-user"
}