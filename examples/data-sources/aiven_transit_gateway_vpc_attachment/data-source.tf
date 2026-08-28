# AWS Transit Gateway
data "aiven_transit_gateway_vpc_attachment" "aws_tgw" {
  vpc_id             = aiven_project_vpc.example.id
  peer_cloud_account = "123456789012"  # AWS account ID
  peer_vpc           = "tgw-0123456789abcdef0"  # AWS Transit Gateway ID
}

# UpCloud VPC peering
data "aiven_transit_gateway_vpc_attachment" "upcloud_peer" {
  vpc_id             = aiven_project_vpc.example.id
  peer_cloud_account = "upcloud"
  peer_vpc           = "03126dc1-a69f-4bc2-8b24-e31c22d64712"  # UpCloud network UUID
}
