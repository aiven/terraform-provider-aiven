# Your Aiven project
data "aiven_project" "main" {
  project = var.aiven_project_name
}

# PostgreSQL service
resource "aiven_pg" "example_pg" {
  project      = data.aiven_project.main.project
  service_name = var.postgres_service_name
  cloud_name   = "aws-eu-west-2"
  plan         = "startup-4"
}

output "pg_service_uri" {
  value     = aiven_pg.example_pg.service_uri
  sensitive = true
}
