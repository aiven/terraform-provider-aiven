terraform {
  required_providers {
    aiven = {
      source = "aiven/aiven"
      version = ">= 2.0.0, < 3.0.0"
    }
  }
}

provider "aiven" {
  api_token = var.aiven_api_token
}
