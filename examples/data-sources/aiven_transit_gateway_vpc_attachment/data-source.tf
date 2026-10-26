data "aiven_transit_gateway_vpc_attachment" "example" {
  project = "my-project"

  // OPTIONAL FIELDS
  project_vpc_id     = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  peer_cloud_account = "123456789012"
  peer_vpc           = "tgw-0123456789abcdef0"
  peer_region        = "us-east-1"

  /* COMPUTED FIELDS
  peering_connection_id = "pcx-0123456789abcdef0"
  state                 = "ACTIVE"
  state_info = {
    foo = "foo"
  }
  user_peer_network_cidrs = ["192.168.6.0/24"]
  */
}
