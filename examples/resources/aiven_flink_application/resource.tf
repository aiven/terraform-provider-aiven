resource "aiven_flink_application" "example_app" {
  project      = data.aiven_project.example_project.project
  service_name = "example-flink-service"
  name         = "example-app"
}

