# S3 bucket that can hold data in formats that ClickHouse can read from,
# eg. CSV, Parquet, ORC, JSON, Avro
resource "aiven_service_integration_endpoint" "s3_bucket" {
  project       = var.aiven_project_name
  endpoint_name = "s3-bucket"
  endpoint_type = "external_aws_s3"

  external_aws_s3_user_config {
    access_key_id     = var.s3_bucket_access_key
    secret_access_key = var.s3_bucket_secret_key
    url               = var.s3_bucket_url
  }
}
