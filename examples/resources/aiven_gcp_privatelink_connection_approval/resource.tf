resource "aiven_gcp_privatelink_connection_approval" "example" {
  project         = "my-project" // Force new
  service_name    = "my-service" // Force new
  user_ip_address = "10.20.30.40"

  // OPTIONAL FIELDS
  psc_connection_id = "foo" // Force new

  /* COMPUTED FIELDS
  id                        = "foo"
  privatelink_connection_id = "foo"
  state                     = "active"
  */
}
