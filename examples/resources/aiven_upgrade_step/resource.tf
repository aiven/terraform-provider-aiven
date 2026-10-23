resource "aiven_upgrade_step" "example" {
  organization_id          = "org1a23f456789" // Force new
  destination_project_name = "prod-project" // Force new
  destination_service_name = "pg-prod" // Force new
  source_project_name      = "dev-project" // Force new
  source_service_name      = "pg-dev" // Force new

  // OPTIONAL FIELDS
  auto_validation_delay_days = 1

  /* COMPUTED FIELDS
  step_id = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  */
}
