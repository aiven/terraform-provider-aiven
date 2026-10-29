resource "aiven_byoc_azure_entity" "example" {
  organization_id  = "org1a23f456789" // Force new
  cloud_region     = "westeurope" // Force new
  deployment_model = "standard" // Force new
  display_name     = "byoc-cloud-prod-westeurope"
  reserved_cidr    = "192.168.6.0/24" // Force new

  // OPTIONAL FIELDS
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
  aiven_azure_cce_client_id    = "foo"
  custom_cloud_environment_id  = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  aiven_management_cidr_blocks = ["10.0.0.0/24"]
  bucket_names = {
    foo = "foo"
  }
  byoc_resource_tags = {
    foo = "foo"
  }
  byoc_unique_name           = "foo"
  use_customer_owned_storage = true
  */
}
