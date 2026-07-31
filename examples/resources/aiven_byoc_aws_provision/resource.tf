resource "aiven_byoc_aws_provision" "example" {
  organization_id             = "org1a23f456789" // Force new
  custom_cloud_environment_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d" // Force new
  aws_iam_role_arn            = "arn:aws:iam::012345678901:root" // Force new

  // OPTIONAL FIELDS
  azure_client_id       = "12345678-1234-1234-1234-123456789abc" // Force new
  azure_subscription_id = "12345678-1234-1234-1234-123456789abc" // Force new
  azure_tenant_id       = "adcf7194-d877-4505-a47a-91fefd96e3b8" // Force new
  azure_client_secret   = "s3cr3t~EXAMPLE_client_secret_value" // Force new

  /* COMPUTED FIELDS
  aiven_aws_assume_role_external_id = "admin"
  aiven_aws_account_principal       = "foo"
  custom_cloud_names                = ["foo"]
  state                             = "active"
  */
}
