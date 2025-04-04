variable "aiven_token" {
  description = "Aiven token"
  type        = string
}

variable "aiven_project_name" {
  description = "Name of an Aiven project assigned to a billing group"
  type        = string
}

variable "valkey_service_name" {
  description = "Name of the Valkey service"
  type        = string
  default     = "example-valkey-service"
}
