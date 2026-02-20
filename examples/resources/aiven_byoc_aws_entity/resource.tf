resource "aiven_byoc_aws_entity" "example" {
  organization_id  = data.aiven_organization.example.id
  cloud_provider   = "aws"
  cloud_region     = "eu-west-1"
  deployment_model = "standard"
  display_name     = "My BYOC Cloud"
  reserved_cidr    = "10.0.0.0/16"

  aws_iam_role_arn = "arn:aws:iam::123456789012:role/aiven-byoc-role"

  contact_emails {
    email = "admin@example.com"
  }

  tags = {
    environment = "production"
  }
}
