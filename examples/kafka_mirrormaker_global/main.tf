terraform {
  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = "2.1.0"
    }
  }
}

# Initialize provider. No other config options than api_token
provider "aiven" {
  api_token = var.aiven_api_token
}

# # Project
# resource "aiven_project" "prj" {
#   project = var.aiven_api_project
#   card_id = var.aiven_card_id
# }
