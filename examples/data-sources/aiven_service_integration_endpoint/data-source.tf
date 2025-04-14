data "aiven_service_integration_endpoint" "example_datadog_endpoint" {
  project       = aiven_project.example_project.project
  endpoint_name = "Datadog endpoint"
}
