data "aiven_gcp_vpc_peering_connection" "example" {
  gcp_project_id = "my-gcp-project"
  vpc_id         = "example-project/example-vpc"
  peer_vpc       = "my-vpc"

  /* COMPUTED FIELDS
  id        = "example-project/example-vpc/my-gcp-project/my-vpc"
  self_link = "https://www.googleapis.com/compute/v1/projects/my-gcp-project/global/networks/my-vpc"
  state     = "ACTIVE"
  state_info = {
    foo = "foo"
  }
  */
}
