terraform {
  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = "2.X.X"
    }
  }
}
provider "aiven" {
  api_token = var.aiven_api_token
}
