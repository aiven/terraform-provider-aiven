data "aiven_gcp_org_vpc_peering_connection" "example" {
  organization_id     = "org1a23f456789"
  organization_vpc_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  gcp_project_id      = "my-gcp-project"
  peer_vpc            = "my-vpc"

  /* COMPUTED FIELDS
  self_link = "https://www.googleapis.com/compute/v1/projects/my-gcp-project/global/networks/my-vpc"
  state     = "ACTIVE"
  */
}
