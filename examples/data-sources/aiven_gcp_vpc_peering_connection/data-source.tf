data "aiven_gcp_vpc_peering_connection" "foo" {
  vpc_id         = data.aiven_project_vpc.vpc.id
  gcp_project_id = "xxxx"
  peer_vpc       = "xxxx"
}
