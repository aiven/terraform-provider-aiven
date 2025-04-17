variable "aiven_token" {
  description = "Aiven token"
  type        = string
  sensitive   = true
}

variable "azure_client_secret" {
  description = "Azure client secret"
  type        = string
  sensitive   = true
}

variable "aiven_project_name" {
  description = "Name of the Aiven project"
  type        = string
}