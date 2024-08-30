data "aiven_m3db_user" "example_service_user" {
  service_name = aiven_m3db.example_m3db.service_name
  project      = data.aiven_project.example_project.project
  username     = "example-m3db-user"
}