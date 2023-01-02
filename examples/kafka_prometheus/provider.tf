terraform {
  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = ">=3.0.0, <4.0.0"
    }
  }
}

# Initialize provider. No other config options than api_token
provider "aiven" {
  api_token = var.avn_token
}
