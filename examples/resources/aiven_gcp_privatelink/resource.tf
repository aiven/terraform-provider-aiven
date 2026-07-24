resource "aiven_gcp_privatelink" "example" {
  project      = "my-project" // Force new
  service_name = "foo" // Force new

  /* COMPUTED FIELDS
  google_service_attachment = "foo"
  state                     = "active"
  */
}
