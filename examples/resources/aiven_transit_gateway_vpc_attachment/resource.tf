resource "aiven_transit_gateway_vpc_attachment" "attachment" {
  vpc_id             = aiven_project_vpc.bar.id
  peer_cloud_account = "<PEER_ACCOUNT_ID>"
  peer_vpc           = "google-project1"
  peer_region        = "aws-eu-west-1"
  user_peer_network_cidrs = [
    "10.0.0.0/24"
  ]
}
