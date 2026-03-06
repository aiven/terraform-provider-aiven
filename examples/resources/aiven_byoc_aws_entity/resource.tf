resource "aiven_byoc_aws_entity" "example" {
  organization_id  = data.aiven_organization.main.id
  display_name     = "my-byoc-cloud"
  cloud_provider   = "aws"
  cloud_region     = "aws-eu-west-1"
  deployment_model = "standard"
  reserved_cidr    = "10.0.0.0/16"
  aws_iam_role_arn = "arn:aws:iam::123456789012:role/my-aiven-byoc-role"

  contact_emails {
    email     = "ops@example.com"
    real_name = "Ops Team"
    role      = "admin"
  }
}
