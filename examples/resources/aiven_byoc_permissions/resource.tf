resource "aiven_byoc_permissions" "example" {
  organization_id             = "org1a23f456789" // Force new
  custom_cloud_environment_id = "foo" // Force new
  accounts                    = ["a22ba494e096"]
  projects                    = ["project-prod"]
}
