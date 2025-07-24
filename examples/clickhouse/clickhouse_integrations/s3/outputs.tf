output "clickhouse_service_uri" {
  description = "ClickHouse service URI"
  value       = aiven_clickhouse.main.service_uri
  sensitive   = true
}

output "clickhouse_service_host" {
  description = "ClickHouse service host"
  value       = aiven_clickhouse.main.service_host
}

output "s3_bucket_name" {
  description = "Name of the S3 bucket"
  value       = aws_s3_bucket.clickhouse_data.bucket
}

output "sample_query" {
  description = "Sample query to test the S3 integration using named collection"
  value       = "SELECT * FROM s3(`endpoint_${aiven_service_integration_endpoint.s3_endpoint.endpoint_name}`, filename='test_data.csv', format='CSVWithNames');"
}

output "users_created" {
  description = "ClickHouse users created with S3 access"
  value = [
    aiven_clickhouse_user.app_user.username,
    aiven_clickhouse_user.demo_user.username
  ]
}

output "endpoint_name" {
  description = "Name of the S3 integration endpoint for use in queries"
  value       = aiven_service_integration_endpoint.s3_endpoint.endpoint_name
}
