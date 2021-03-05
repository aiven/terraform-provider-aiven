terraform {
  required_providers {
    aiven = {
            source = "aiven/aiven"
            version = "2.X.X"
    }
  }
}

# Initialize provider. No other config options than api_token
provider "aiven" {
    api_token = var.aiven_api_token
}

