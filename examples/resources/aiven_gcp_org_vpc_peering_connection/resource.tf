resource "aiven_gcp_org_vpc_peering_connection" "example" {
  organization_id     = "org1a23f456789" // Force new
  organization_vpc_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d" // Force new
  gcp_project_id      = "my-gcp-project" // Force new
  peer_vpc            = "my-vpc" // Force new

  /* COMPUTED FIELDS
  self_link = "https://www.googleapis.com/compute/v1/projects/my-gcp-project/global/networks/my-vpc"
  state     = "ACTIVE"
  */
}
