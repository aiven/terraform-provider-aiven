resource "aiven_organization_vpc" "example_vpc" {
  organization_id = data.aiven_organization.example.id
  cloud_name      = "google-europe-west10"
  network_cidr    = "10.0.0.0/24"
}

resource "aiven_gcp_org_vpc_peering_connection" "example" {
  organization_id     = aiven_organization_vpc.example_vpc.organization_id
  organization_vpc_id = aiven_organization_vpc.example_vpc.organization_vpc_id
  gcp_project_id      = "my-gcp-project-123" # Your GCP project ID
  peer_vpc            = "my-vpc-network"     # Your GCP VPC network name
}
