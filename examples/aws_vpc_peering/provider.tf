terraform {
  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = ">= 4.0.0, < 5.0.0"
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
  region = "ap-southeast-2"
}
