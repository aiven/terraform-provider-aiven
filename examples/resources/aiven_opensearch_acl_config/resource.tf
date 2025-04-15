resource "aiven_opensearch_acl_config" "main" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_opensearch.example_opensearch.service_name
  enabled      = true
  extended_acl = false
}
