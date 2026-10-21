resource "aiven_transit_gateway_vpc_attachment" "example" {
  peer_cloud_account = "123456789012" // Force new
  peer_vpc           = "tgw-0123456789abcdef0" // Force new

  // OPTIONAL FIELDS
  project                 = "example-project"
  project_vpc_id          = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  peer_region             = "us-east-1" // Force new
  user_peer_network_cidrs = ["192.168.6.0/24"]

  /* COMPUTED FIELDS
  peering_connection_id = "pcx-0123456789abcdef0"
  state                 = "ACTIVE"
  state_info = {
    foo = "foo"
  }
  */
}
