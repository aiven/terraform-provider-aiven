# Initialize the provider
terraform {
  required_version = ">=0.13"
  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = ">=4.0.0, <5.0.0"
    }
  }
}

provider "aiven" {
  api_token = var.aiven_api_token
}

variable "aiven_api_token" {
  description = "Aiven API token"
  type        = string
}
