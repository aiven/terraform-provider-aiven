terraform {
  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = ">= 2.0.0, < 3.0.0"
    }
  }
}

# Initialize provider. No other config options than api_token
provider "aiven" {
}
