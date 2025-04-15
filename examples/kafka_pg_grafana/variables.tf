variable "aiven_token" {
  description = "Aiven token"
  type        = string
  sensitive   = true
}

variable "aiven_project_name" {
  description = "Name of an Aiven project assigned to a billing group"
  type        = string
}

variable "postgres_service_name" {
  description = "Name of the PostgreSQL service"
  type        = string
  default     = "example-pg-service"
}

variable "kafka_service_name" {
  description = "Name of the Kafka service"
  type        = string
  default     = "example-kafka-service"
}

variable "grafana_service_name" {
  description = "Name of the Grafana service"
  type        = string
  default     = "example-grafana-service"
}