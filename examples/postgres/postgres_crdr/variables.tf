variable "aiven_token" {
  description = "Aiven token"
  type        = string
}

variable "aiven_project_name" {
  description = "Name of your Aiven project"
  type        = string
}

variable "primary_service_name" {
  description = "Name of the primary PostgreSQL service"
  type        = string
  default     = "example-pg-primary"
}

variable "recovery_service_name" {
  description = "Name of the recovery PostgreSQL service"
  type        = string
  default     = "example-pg-recovery"
}

variable "service_plan" {
  description = "Service plan"
  type        = string
  default     = "startup-4"
}

variable "primary_cloud_name" {
  description = "Cloud region of the primary service"
  type        = string
  default     = "google-europe-west1"
}

variable "recovery_cloud_name" {
  description = "Cloud region of the recovery service"
  type        = string
  default     = "google-europe-west4"
}
