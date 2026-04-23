resource "aiven_byoc_aws_provision" "example" {
  organization_id             = "org1a23f456789" // Force new
  custom_cloud_environment_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d" // Force new
  aws_iam_role_arn            = "arn:aws:iam::012345678901:root" // Force new

  /* COMPUTED FIELDS
  aiven_aws_assume_role_external_id = "admin"
  aiven_aws_account_principal       = "foo"
  */
}
