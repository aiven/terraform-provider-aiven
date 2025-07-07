variable "aiven_token" {
  description = "Aiven token"
  type        = string
}

variable "aiven_project_name" {
  description = "Name of an Aiven project assigned to a billing group"
  type        = string
}

variable "kafka_service_name" {
  description = "Name of the Kafka service"
  type        = string
  default     = "example-kafka-service"
}

variable "kafka_connect_name" {
  description = "Name of the Kafka Connect service"
  type        = string
  default     = "example-kafka-connect-service"
}
