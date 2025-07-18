output "clickhouse_service_host" {
  value = aiven_clickhouse.dev.service_host
}

output "clickhouse_service_port" {
  value = aiven_clickhouse.dev.service_port
}

output "clickhouse_service_username" {
  value = aiven_clickhouse.dev.service_username
}

output "clickhouse_service_password" {
  value     = aiven_clickhouse.dev.service_password
  sensitive = true
}

locals {
  https_component = [for c in aiven_clickhouse.dev.components : c if c.component == "clickhouse_https"][0]
}

output "clickhouse_https_uri" {
  description = "HTTPS connection URI for ClickHouse"
  value       = "https://${aiven_clickhouse.dev.service_username}:${aiven_clickhouse.dev.service_password}@${local.https_component.host}:${local.https_component.port}"
  sensitive   = true
}
