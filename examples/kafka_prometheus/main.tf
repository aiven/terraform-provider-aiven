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

data "aiven_project" "bd" {
  project = var.project
}


resource "aiven_kafka" "bar" {
  project                 = data.aiven_project.bd.project
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  service_name            = var.kafka_svc
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  kafka_user_config {
    kafka_rest      = true
    kafka_connect   = true
    schema_registry = true
    kafka_version   = "2.6"
    ip_filter       = ["0.0.0.0/0", "80.242.179.94", "188.166.141.226"]
    kafka {
      group_max_session_timeout_ms = 70000
      log_retention_bytes          = 1000000000
    }
    public_access {
      kafka_rest    = true
      kafka_connect = true
      prometheus    = true
    }
  }
}

resource "aiven_service_integration_endpoint" "prom" {
  project       = data.aiven_project.bd.project
  endpoint_name = var.prom_name
  endpoint_type = "prometheus"
  prometheus_user_config {
    basic_auth_username = "jfklhjfgfgfsgf"
    basic_auth_password = "fjfdkljdfdsgfgfgf"
  }
}

resource "aiven_service_integration" "rsys_int" {
  project                 = data.aiven_project.bd.project
  destination_endpoint_id = aiven_service_integration_endpoint.prom.id
  integration_type        = "prometheus"
  source_service_name     = aiven_kafka.bar.service_name
}
