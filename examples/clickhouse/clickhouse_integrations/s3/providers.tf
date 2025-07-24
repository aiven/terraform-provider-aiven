terraform {
  required_version = ">=0.13"

  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = ">=4.40.0, <5.0.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aiven" {
  api_token = var.aiven_token
}

provider "aws" {
  region     = var.aws_region
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
}
