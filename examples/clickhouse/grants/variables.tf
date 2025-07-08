variable "aiven_token" {
  description = "Aiven authentication token"
  type        = string
  sensitive   = true
}

variable "aiven_project" {
  description = "Aiven project name where the ClickHouse service will be created"
  type        = string
  default     = "demo-project"
}

variable "cloud" {
  description = "Cloud provider and region for the ClickHouse service"
  type        = string
  default     = "google-europe-north1"
}

variable "plan" {
  description = "Aiven service plan (startup-4, startup-8, startup-16, business-4, etc.)"
  type        = string
  default     = "startup-16"
}

variable "databases" {
  description = "List of databases to create in ClickHouse service"
  type        = list(string)
  default = [
    "main_db",
    "test_db",
    "staging_db",
  ]
}
