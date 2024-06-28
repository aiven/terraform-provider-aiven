data "aiven_flink" "example_flink" {
  project      = data.aiven_project.example_project.project
  service_name = "example-flink-service"
}
