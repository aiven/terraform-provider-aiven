resource "aiven_organization_vpc" "example_vpc" {
  organization_id = data.aiven_organization.example.id
  cloud_name     = "aws-eu-central-1"
  network_cidr   = "10.0.0.0/24"
}

resource "aiven_aws_org_vpc_peering_connection" "example_peering" {
  organization_id         = aiven_organization_vpc.example_vpc.organization_id
  organization_vpc_id         = aiven_organization_vpc.example_vpc.organization_vpc_id
  aws_account_id = var.aws_id
  aws_vpc_id     = "vpc-1a2b3c4d5e6f7g8h9"
  aws_vpc_region = "aws-us-east-2"
}
