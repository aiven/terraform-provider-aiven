# AWS Transit Gateway attachment
resource "aiven_transit_gateway_vpc_attachment" "aws_tgw" {
  vpc_id             = aiven_project_vpc.example.id
  peer_cloud_account = "123456789012"          # AWS account ID
  peer_vpc           = "tgw-0123456789abcdef0" # AWS Transit Gateway ID
  peer_region        = "eu-west-1"
  user_peer_network_cidrs = [
    "10.0.0.0/24"
  ]
}

# UpCloud VPC peering attachment
resource "aiven_transit_gateway_vpc_attachment" "upcloud_peer" {
  vpc_id             = aiven_project_vpc.example.id
  peer_cloud_account = "upcloud"
  peer_vpc           = "03126dc1-a69f-4bc2-8b24-e31c22d64712" # UpCloud network UUID
  user_peer_network_cidrs = [
    "192.168.0.0/24"
  ]
}
