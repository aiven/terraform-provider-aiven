data "aiven_opensearch" "example_opensearch" {
  project      = data.aiven_project.example_project.project
  service_name = "example-opensearch"
}
