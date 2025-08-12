variable "aiven_token" {
  description = "Aiven token"
  type        = string
}

variable "project_name" {
  description = "Name of the Aiven project"
  type        = string
}

variable "postgresql_service_name" {
  description = "Name of the PostgreSQL service"
  type        = string
  default     = "demo-postgres"
}

variable "kafka_service_name" {
  description = "Name of the Kafka service"
  type        = string
  default     = "demo-kafka"
}

variable "kafka_connect_service_name" {
  description = "Name of the Kafka Connect service"
  type        = string
  default     = "demo-kafka-connect"
}
