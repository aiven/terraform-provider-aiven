data "aiven_gcp_vpc_peering_connection" "main" {
  vpc_id         = data.aiven_project_vpc.vpc.id
  gcp_project_id = "example-project"
  peer_vpc       = "example-network"
}
