resource "aiven_static_ip" "example" {
  project    = "my-project" // Force new
  cloud_name = "aws-eu-central-1" // Force new

  // OPTIONAL FIELDS
  termination_protection = false

  /* COMPUTED FIELDS
  static_ip_address_id = "192.168.1.1"
  ip_address           = "foo"
  service_name         = "foo"
  state                = "assigned"
  */
}
