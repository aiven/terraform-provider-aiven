resource "aiven_clickhouse" "clickhouse" {
  project      = var.aiven_project_name
  cloud_name   = "google-europe-west1"
  plan         = "startup-16"
  service_name = "clickhouse-gcp-eu"
}

# Second ClickHouse service, used to showcase integrating with an external cluster

resource "aiven_clickhouse" "external_clickhouse" {
  project      = var.aiven_project_name
  cloud_name   = "google-europe-west1"
  plan         = "startup-16"
  service_name = "external-clickhouse-gcp-eu"
}

resource "aiven_service_integration_endpoint" "external_clickhouse" {
  project       = var.aiven_project_name
  endpoint_name = "external-clickhouse"
  endpoint_type = "external_clickhouse"

  external_clickhouse_user_config {
    host     = aiven_clickhouse.external_clickhouse.service_host
    port     = aiven_clickhouse.external_clickhouse.service_port
    username = aiven_clickhouse.external_clickhouse.service_username
    password = aiven_clickhouse.external_clickhouse.service_password
  }
}
