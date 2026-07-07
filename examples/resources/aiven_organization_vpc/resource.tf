resource "aiven_organization_vpc" "example" {
  organization_id = "org1a23f456789" // Force new
  cloud_name      = "aws-eu-west-1" // Force new
  network_cidr    = "10.0.0.0/24" // Force new

  // OPTIONAL FIELDS
  display_name = "My organization VPC"

  /* COMPUTED FIELDS
  organization_vpc_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  create_time         = "2021-01-01T00:00:00Z"
  state               = "ACTIVE"
  update_time         = "2021-01-01T00:00:00Z"
  */
}
