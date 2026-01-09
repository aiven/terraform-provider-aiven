variable "aiven_token" {
  description = "Aiven token"
  type        = string
  sensitive   = true
}

variable "aiven_project" {
  description = "Aiven project name"
  type        = string
}

variable "cloud" {
  description = "Cloud provider and region"
  type        = string
  default     = "google-europe-west1"
}

variable "plan" {
  description = "Service plan name"
  type        = string
  default     = "business-4"
}
variable "kafka_service_name" {
  description = "Name of the Kafka service"
  type        = string
  default     = "example-kafka-sasl-enabled"
}
