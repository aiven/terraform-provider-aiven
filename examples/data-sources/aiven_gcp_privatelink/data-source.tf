data "aiven_gcp_privatelink" "example" {
  project      = "my-project"
  service_name = "foo"

  /* COMPUTED FIELDS
  google_service_attachment = "foo"
  state                     = "active"
  */
}
