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
