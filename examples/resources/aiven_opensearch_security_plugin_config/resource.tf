resource "aiven_opensearch_security_plugin_config" "example" {
  project        = "my-project" // Force new
  service_name   = "my-opensearch" // Force new
  admin_password = "h3.2aD!z2"

  /* COMPUTED FIELDS
  admin_enabled = true
  available     = true
  enabled       = true
  */
}
