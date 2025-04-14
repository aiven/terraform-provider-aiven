data "aiven_valkey" "example_valkey" {
  project      = data.aiven_project.example_project.project
  service_name = "example-valkey-service"
}
