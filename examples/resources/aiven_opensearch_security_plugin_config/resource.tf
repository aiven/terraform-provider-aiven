resource "aiven_opensearch_security_plugin_config" "main" {
  project        = data.aiven_project.example_project.project
  service_name   = aiven_opensearch.example_opensearch.service_name
  admin_password = var.opensearch_security_admin_password
}
