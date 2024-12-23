terraform {
  required_version = ">=0.13"
  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = ">=4.0.0, <5.0.0"
    }
  }
}

variable "aiven_api_token" {
  type = string
}

variable "aiven_card_id" {
  type = string
}

provider "aiven" {
  api_token = var.aiven_api_token
}

resource "aiven_project" "project" {
  project = "static-ips-project"
}

resource "aiven_static_ip" "ips" {
  count = 6

  project    = aiven_project.project.project
  cloud_name = "google-europe-west1"
}

resource "aiven_pg" "pg" {
  project      = aiven_project.project.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "pg-with-static-ips"

  static_ips = toset([
    aiven_static_ip.ips[0].static_ip_address_id,
    aiven_static_ip.ips[1].static_ip_address_id,
    aiven_static_ip.ips[2].static_ip_address_id,
    aiven_static_ip.ips[3].static_ip_address_id,
    aiven_static_ip.ips[4].static_ip_address_id,
    aiven_static_ip.ips[5].static_ip_address_id,
  ])

  pg_user_config {
    static_ips = true
  }
}
