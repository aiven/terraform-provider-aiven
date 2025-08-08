variable "aiven_token" {
  description = "Aiven token"
  type        = string
}

variable "aiven_project" {
  description = "Aiven project name"
  type        = string
}

variable "kafka_service_name" {
  description = "Name of the Kafka service"
  type        = string
  default     = "example-kafka-with-karapace-schema-registry"
}
