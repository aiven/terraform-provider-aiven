data "aiven_m3db" "example_m3db" {
  project      = data.aiven_project.example_project.project
  service_name = "example-m3db-service"
}

