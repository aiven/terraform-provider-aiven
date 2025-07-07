variable "aiven_token" {
  description = "Aiven API token"
  type        = string
  sensitive   = true
}

variable "aiven_project_name" {
  description = "Aiven project name"
  type        = string
}

variable "clickhouse_service_name" {
  description = "Name of the ClickHouse service"
  type        = string
}

variable "external_clickhouse_service_name" {
  description = "Name of the external ClickHouse service"
  type        = string
}

variable "external_mysql_service_name" {
  description = "Name of the external MySQL service"
  type        = string
}

variable "external_postgres_service_name" {
  description = "Name of the external Postgres service"
  type        = string
}

variable "s3_bucket_access_key" {
  default = "AKIAIOSFODNN7EXAMPLE"
  type    = string
}

variable "s3_bucket_secret_key" {
  default   = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  type      = string
  sensitive = true
}

variable "s3_bucket_url" {
  default = "https://mybucket.s3-myregion.amazonaws.com/mydataset/"
  type    = string
}
