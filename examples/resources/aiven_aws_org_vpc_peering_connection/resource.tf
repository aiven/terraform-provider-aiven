resource "aiven_aws_org_vpc_peering_connection" "example" {
  organization_id     = "org1a23f456789" // Force new
  organization_vpc_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d" // Force new
  aws_account_id      = "123456789012" // Force new
  aws_vpc_id          = "vpc-2f09a348" // Force new
  aws_vpc_region      = "us-east-1" // Force new

  /* COMPUTED FIELDS
  aws_vpc_peering_connection_id = "pcx-1234567890abcdef0"
  peering_connection_id         = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  state                         = "ACTIVE"
  */
}
