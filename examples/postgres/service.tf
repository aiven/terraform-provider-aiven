# European Postgres Service
resource "aiven_pg" "avn-eu-pg" {
  project      = var.avn_project_id
  cloud_name   = "aws-eu-west-2" # London
  plan         = "business-8"    # Primary + read replica
  service_name = "postgres-eu"
}

# US Postgres Service
resource "aiven_pg" "avn-us-pg" {
  project      = var.avn_project_id
  cloud_name   = "do-nyc"     # New York
  plan         = "business-8" # Primary + read replica
  service_name = "postgres-us"
}

# Asia Postgres Service
resource "aiven_pg" "avn-as-pg" {
  project      = var.avn_project_id
  cloud_name   = "google-asia-southeast1" # Singapore
  plan         = "business-8"             # Primary + read replica
  service_name = "postgres-as"
}
