# Your Aiven project
data "aiven_project" "example_project" {
  project = var.aiven_project_name
}

# Create a VPC in AWS
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  tags = {
    Name = "example-vpc"
  }
}

# Create a subnet in the AWS VPC
resource "aws_subnet" "main" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = "example-subnet"
  }
}

# Create a VPC in your Aiven project
resource "aiven_project_vpc" "main" {
  project      = data.aiven_project.example_project.project
  cloud_name   = "aws-ap-southeast-2"
  network_cidr = "192.168.0.0/24" # Ensure the CIDR range of this VPC doesn't overlap with the range of your AWS VPC.
}

# Create a VPC peering connection between your Aiven and AWS VPCs
resource "aiven_aws_vpc_peering_connection" "peer_to_aws" {
  vpc_id         = aiven_project_vpc.main.id
  aws_account_id = var.aws_account_id
  aws_vpc_id     = aws_vpc.main.id
  aws_vpc_region = "ap-southeast-2"
  depends_on = [
    aiven_project_vpc.main, aws_vpc.main
  ]
}

# Accept the VPC peering request from Aiven
resource "aws_vpc_peering_connection_accepter" "peer" {
  vpc_peering_connection_id = aiven_aws_vpc_peering_connection.peer_to_aws.aws_vpc_peering_connection_id
  auto_accept               = true

  tags = {
    Side = "Accepter"
  }

  depends_on = [
    aiven_aws_vpc_peering_connection.peer_to_aws
  ]
}

# Update route tables, routing the Aiven VPC CIDR through the peering connection
resource "aws_route_table" "route_aiven" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block                = "192.168.0.0/24"
    vpc_peering_connection_id = aiven_aws_vpc_peering_connection.peer_to_aws.aws_vpc_peering_connection_id
  }
}

# Associate the route table with the subnet
resource "aws_route_table_association" "subnet_aiven" {
  subnet_id      = aws_subnet.main.id
  route_table_id = aws_route_table.route_aiven.id
}
