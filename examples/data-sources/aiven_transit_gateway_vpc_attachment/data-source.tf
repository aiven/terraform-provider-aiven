data "aiven_transit_gateway_vpc_attachment" "example" {
  vpc_id             = "example-project/example-vpc"
  peer_cloud_account = "123456789012"
  peer_vpc           = "tgw-0123456789abcdef0"

  // OPTIONAL FIELDS
  peer_region = "us-east-1"

  /* COMPUTED FIELDS
  id                    = "example-project/example-vpc/123456789012/tgw-0123456789abcdef0/us-east-1"
  peering_connection_id = "pcx-0123456789abcdef0"
  state                 = "ACTIVE"
  state_info = {
    foo = "foo"
  }
  user_peer_network_cidrs = ["192.168.6.0/24"]
  */
}
