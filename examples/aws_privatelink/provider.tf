terraform {
  required_providers {
    aiven = {
      //      source = "aiven/aiven"
      //      version = "2.X.X"
      source = "aiven.io/provider/aiven"
      version = "2.0.0"
    }
  }
}

# Initialize provider. No other config options than api_token
provider "aiven" {
  //  api_token = var.aiven_api_token
}

