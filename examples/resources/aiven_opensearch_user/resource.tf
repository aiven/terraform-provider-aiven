resource "aiven_opensearch_user" "example_opensearch_user" {
  service_name = aiven_opensearch.example_opensearch.service_name
  project      = data.aiven_project.example_project.project
  username     = "example-opensearch-user"
  password     = var.opensearch_user_password
}
