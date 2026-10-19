resource "aiven_aws_vpc_peering_connection" "example" {
  aws_account_id = "123456789012" // Force new
  aws_vpc_id     = "vpc-0123456789abcdef0" // Force new
  aws_vpc_region = "eu-west-1"

  // OPTIONAL FIELDS
  project        = "example-project"
  project_vpc_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"

  /* COMPUTED FIELDS
  aws_vpc_peering_connection_id = "pcx-0123456789abcdef0"
  create_time                   = "2021-01-01T00:00:00Z"
  state                         = "ACTIVE"
  state_info = {
    foo = "foo"
  }
  */
}
