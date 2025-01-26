data "aiven_mysql" "example_mysql" {
  project      = aiven_project.example_project.project
  service_name = "example-mysql"
}
