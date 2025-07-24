variable "aiven_token" {
  description = "Aiven authentication token"
  type        = string
  sensitive   = true
}

variable "aws_access_key" {
  description = "AWS access key"
  type        = string
  sensitive   = true
}

variable "aws_secret_key" {
  description = "AWS secret key"
  type        = string
  sensitive   = true
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "eu-central-1"
}

variable "aiven_project" {
  description = "Aiven project name"
  type        = string
}

variable "cloud" {
  description = "Cloud provider and region"
  type        = string
  default     = "google-europe-north1"
}

variable "plan" {
  description = "Aiven service plan"
  type        = string
  default     = "startup-16"
}

variable "service_name_prefix" {
  description = "Prefix for service and resource names"
  type        = string
  default     = "clickhouse-s3"
}
