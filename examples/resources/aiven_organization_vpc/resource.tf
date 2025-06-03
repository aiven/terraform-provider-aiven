resource "aiven_organization_vpc" "example_vpc" {
  organization_id = data.aiven_organization.example.id
  cloud_name      = "aws-eu-central-1"
  network_cidr    = "10.0.0.0/24"
}
