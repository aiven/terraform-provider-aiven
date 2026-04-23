resource "aiven_byoc_aws_entity" "example" {
  organization_id  = "org1a23f456789" // Force new
  cloud_provider   = "aws" // Force new
  cloud_region     = "eu-west-1" // Force new
  deployment_model = "standard" // Force new
  display_name     = "byoc-cloud-prod-eu-west-1"
  reserved_cidr    = "192.168.6.0/24" // Force new

  // OPTIONAL FIELDS
  aws_iam_role_arn = "arn:aws:iam::012345678901:root"
  contact_emails {
    email = "jane@example.com"

    // OPTIONAL FIELDS
    real_name = "Jane Smith"
    role      = "admin"
  }
  tags = {
    foo = "foo"
  }

  /* COMPUTED FIELDS
  aiven_aws_assume_role_external_id                = "admin"
  custom_cloud_environment_id                      = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  aiven_aws_account_principal                      = "foo"
  aiven_aws_object_storage_credentials_creator_arn = "foo"
  aiven_aws_object_storage_user_arn                = "foo"
  aiven_management_cidr_blocks                     = ["10.0.0.0/24"]
  aiven_object_storage_credentials_creator_user    = "foo"
  aws_subnets_bastion = {
    foo = "foo"
  }
  aws_subnets_workload = {
    foo = "foo"
  }
  bucket_names = {
    foo = "foo"
  }
  byoc_resource_tags = {
    foo = "foo"
  }
  byoc_unique_name   = "foo"
  custom_cloud_names = ["foo"]
  errors {
    category = "general_error"
    message  = "foo"
  }
  state                      = "active"
  update_time                = "2021-01-01T00:00:00Z"
  use_customer_owned_storage = true
  */
}
