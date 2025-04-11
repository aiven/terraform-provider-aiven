data "aiven_opensearch_acl_config" "os-acl-config" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_opensearch.example_opensearch.service_name
}

