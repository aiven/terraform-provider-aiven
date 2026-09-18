resource "aiven_aws_vpc_peering_connection" "example" {
  aws_account_id = "123456789012" // Force new
  aws_vpc_id     = "vpc-0123456789abcdef0" // Force new
  vpc_id         = "example-project/example-vpc" // Force new
  aws_vpc_region = "eu-west-1"

  /* COMPUTED FIELDS
  aws_vpc_peering_connection_id = "pcx-0123456789abcdef0"
  id                            = "example-project/example-vpc/123456789012/vpc-0123456789abcdef0/eu-west-1"
  state                         = "ACTIVE"
  state_info = {
    foo = "foo"
  }
  */
}
