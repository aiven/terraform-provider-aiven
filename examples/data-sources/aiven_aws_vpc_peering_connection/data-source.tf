resource "aiven_project_vpc" "example_vpc" {
  project      = data.aiven_project.example_project.project
  cloud_name   = "google-europe-west1"
  network_cidr = "192.168.1.0/24"
}

data "aiven_aws_vpc_peering_connection" "aws_to_aiven_peering" {
  vpc_id         = aiven_project_vpc.example_vpc.id
  aws_account_id = var.aws_id
  aws_vpc_id     = "vpc-1a2b3c4d5e6f7g8h9"
  aws_vpc_region = "aws-us-east-2"
}
