# ClickHouse credentials integrations with endpoints defined in clickhouse.tf, mysql.tf and postgres.tf
# Each integration will result in a new named collection being created in the ClickHouse service.

resource "aiven_service_integration" "s3_managed_credentials" {
  project                  = var.aiven_project_name
  integration_type         = "clickhouse_credentials"
  source_endpoint_id       = aiven_service_integration_endpoint.s3_bucket.id
  destination_service_name = aiven_clickhouse.clickhouse.service_name
}

resource "aiven_service_integration" "external_postgres_managed_credentials" {
  project                  = var.aiven_project_name
  integration_type         = "clickhouse_credentials"
  source_endpoint_id       = aiven_service_integration_endpoint.external_postgres.id
  destination_service_name = aiven_clickhouse.clickhouse.service_name
}

resource "aiven_service_integration" "external_mysql_managed_credentials" {
  project                  = var.aiven_project_name
  integration_type         = "clickhouse_credentials"
  source_endpoint_id       = aiven_service_integration_endpoint.external_mysql.id
  destination_service_name = aiven_clickhouse.clickhouse.service_name
}


resource "aiven_service_integration" "external_clickhouse_managed_credentials" {
  project                  = var.aiven_project_name
  integration_type         = "clickhouse_credentials"
  source_endpoint_id       = aiven_service_integration_endpoint.external_clickhouse.id
  destination_service_name = aiven_clickhouse.clickhouse.service_name
}

