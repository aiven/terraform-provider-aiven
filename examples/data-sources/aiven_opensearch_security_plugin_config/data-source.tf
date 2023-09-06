data "aiven_opensearch_security_plugin_config" "os-sec-config" {
  project      = aiven_project.os-project.project
  service_name = aiven_opensearch.os.service_name
}
