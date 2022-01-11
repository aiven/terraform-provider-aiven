data "aiven_mysql" "mysql1" {
  project      = data.aiven_project.foo.project
  service_name = "my-mysql1"
}
