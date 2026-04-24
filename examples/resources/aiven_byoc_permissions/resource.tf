resource "aiven_byoc_permissions" "example" {
  organization_id             = "org1a23f456789" // Force new
  custom_cloud_environment_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d" // Force new
  accounts                    = ["a22ba494e096"]
  projects                    = ["project-prod"]
}
