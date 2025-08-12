resource "aiven_pg" "postgres" {
  project      = var.project_name
  service_name = var.postgresql_service_name
  cloud_name   = "azure-norway-east"
  plan         = "business-4"
}
