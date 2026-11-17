# Primary PostgreSQL service
resource "aiven_pg" "primary" {
  project      = var.aiven_project_name
  service_name = var.primary_service_name
  cloud_name   = var.primary_cloud_name
  plan         = var.service_plan
}

# Recovery PostgreSQL service in another region.
resource "aiven_pg" "recovery" {
  project      = var.aiven_project_name
  service_name = var.recovery_service_name
  cloud_name   = var.recovery_cloud_name
  plan         = var.service_plan

  service_integrations {
    integration_type    = "disaster_recovery"
    source_service_name = aiven_pg.primary.service_name
  }
}
