resource "aiven_aws_vpc_peering_connection" "foo" {
  vpc_id         = data.aiven_project_vpc.vpc.id
  aws_account_id = "XXXXX"
  aws_vpc_id     = "XXXXX"
}
