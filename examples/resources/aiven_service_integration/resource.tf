resource "aiven_service_integration" "myintegration" {
  project = aiven_project.myproject.project
  // use destination_endpoint_id or destination_service_name = "aiven_service.YYY.service_name"
  destination_endpoint_id = aiven_service_integration_endpoint.XX.id
  integration_type        = "datadog"
  // use source_service_name or source_endpoint_id = aiven_service_integration_endpoint.XXX.id
  source_service_name = aiven_kafka.XXX.service_name
}

