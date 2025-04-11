data "aiven_opensearch_security_plugin_config" "os-sec-config" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_opensearch.example_opensearch.service_name
}
