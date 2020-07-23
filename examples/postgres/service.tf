# European Postgres Service
resource "aiven_service" "avn-eu-pg" {
  project      = var.avn_project_id
  cloud_name   = "aws-eu-west-2" # London
  plan         = "business-8"    # Primary + read replica
  service_name = "postgres-eu"
  service_type = "pg"
}

# US Postgres Service
resource "aiven_service" "avn-us-pg" {
  project      = var.avn_project_id
  cloud_name   = "do-nyc"     # New York
  plan         = "business-8" # Primary + read replica
  service_name = "postgres-us"
  service_type = "pg"
}

# Asia Postgres Service
resource "aiven_service" "avn-as-pg" {
  project      = var.avn_project_id
  cloud_name   = "google-asia-southeast1" # Singapore
  plan         = "business-8"             # Primary + read replica
  service_name = "postgres-as"
  service_type = "pg"
}
