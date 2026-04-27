resource "aiven_project_vpc" "example" {
  project      = "my-project" // Force new
  cloud_name   = "aws-eu-central-1" // Force new
  network_cidr = "192.168.6.0/24" // Force new

  /* COMPUTED FIELDS
  project_vpc_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  state          = "ACTIVE"
  */
}
