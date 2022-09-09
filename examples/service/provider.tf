terraform {
  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = ">=3.0.0"
    }
  }
}

provider "aiven" {
  api_token = var.aiven_api_token
}
