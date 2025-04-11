data "aiven_opensearch_user" "example_opensearch_user" {
  service_name = "example-opensearch-service"
  project      = data.aiven_project.example_project.project
  username     = "example-opensearch-user"
}