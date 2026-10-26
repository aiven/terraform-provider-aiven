resource "aiven_gcp_vpc_peering_connection" "example" {
  gcp_project_id = "my-gcp-project" // Force new
  peer_vpc       = "my-vpc" // Force new

  // OPTIONAL FIELDS
  project_vpc_id = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  project        = "example-project"

  /* COMPUTED FIELDS
  id          = "example-project/1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d/my-gcp-project/my-vpc"
  create_time = "2021-01-01T00:00:00Z"
  self_link   = "https://www.googleapis.com/compute/v1/projects/my-gcp-project/global/networks/my-vpc"
  state       = "ACTIVE"
  state_info = {
    foo = "foo"
  }
  */
}
