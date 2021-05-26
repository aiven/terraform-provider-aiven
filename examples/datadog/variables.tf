variable "aiven_api_token" {
  sensitive = true
  type      = string
}

variable "datadog_api_key" {
  description = "API Key for the Datadog Agent to submit metrics to Datadog"
  sensitive   = true
  type        = string
}
variable "datadog_site" {
  default = "datadoghq.eu"
  type    = string
}
