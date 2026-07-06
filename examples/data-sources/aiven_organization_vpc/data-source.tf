data "aiven_organization_vpc" "example" {
  organization_id     = "org1a23f456789"
  organization_vpc_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"

  /* COMPUTED FIELDS
  cloud_name   = "aws-eu-west-1"
  create_time  = "2021-01-01T00:00:00Z"
  display_name = "My organization VPC"
  network_cidr = "10.0.0.0/24"
  state        = "ACTIVE"
  update_time  = "2021-01-01T00:00:00Z"
  */
}
