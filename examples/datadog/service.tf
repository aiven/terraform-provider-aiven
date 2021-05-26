data "aiven_project" "project" {
  project = "datadog-project"
}

resource "aiven_service_integration_endpoint" "datadog" {
  project       = data.aiven_project.project.project
  endpoint_name = "Datadog"
  endpoint_type = "datadog"

  datadog_user_config {
    datadog_api_key = var.datadog_api_key
    datadog_tags {
      tag = "foo:bar"
    }
    datadog_tags {
      tag = "foo:baz"
    }
    site = var.datadog_site
  }
}

resource "aiven_kafka" "service" {
  project      = data.aiven_project.project.project
  cloud_name   = "google-europe-west1"
  plan         = "business-4"
  service_name = "kafka"
}

resource "aiven_service_integration" "service-integration" {
  project                 = aiven_kafka.service.project
  destination_endpoint_id = aiven_service_integration_endpoint.datadog.id
  integration_type        = "datadog"
  source_service_name     = aiven_kafka.service.service_name
}
