# European Postgres Service
resource "aiven_pg" "avn-eu-pg" {
  project      = var.avn_project
  service_name = var.postgres_eu_name
  cloud_name   = "aws-eu-west-2" # London
  plan         = "startup-4"
}

# US Postgres Service
resource "aiven_pg" "avn-us-pg" {
  project      = var.avn_project
  service_name = var.postgres_us_name
  cloud_name   = "do-nyc"     # New York
  plan         = "business-8" # Primary + read replica
}

# Asia Postgres Service
resource "aiven_pg" "avn-as-pg" {
  project      = var.avn_project
  service_name = var.postgres_as_name
  cloud_name   = "google-asia-southeast1" # Singapore
  plan         = "business-8"             # Primary + read replica
}
