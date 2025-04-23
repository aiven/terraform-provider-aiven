data "aiven_transit_gateway_vpc_attachment" "attachment" {
  vpc_id             = aiven_project_vpc.bar.id
  peer_cloud_account = "<PEER_ACCOUNT_ID>"
  peer_vpc           = "google-project1"
}
