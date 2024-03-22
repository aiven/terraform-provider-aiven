data "aiven_dragonfly" "example_dragonfly" {
  project      = data.aiven_project.example_project.project
  service_name = "example-dragonfly-service"
}
