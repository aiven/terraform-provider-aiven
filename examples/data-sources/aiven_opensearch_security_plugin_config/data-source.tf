data "aiven_opensearch_security_plugin_config" "example" {
  project      = "my-project"
  service_name = "my-opensearch"

  /* COMPUTED FIELDS
  admin_enabled  = true
  admin_password = "h3.2aD!z2"
  available      = true
  enabled        = true
  */
}
