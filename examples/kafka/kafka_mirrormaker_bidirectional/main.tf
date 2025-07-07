terraform {
  required_version = ">=0.13"
  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = ">=4.0.0, <5.0.0"
    }
  }
}

# Initialize provider. No other config options than api_token
provider "aiven" {
  api_token = var.aiven_api_token
}
