data "aiven_service_integration" "myintegration" {
  project                  = aiven_project.myproject.project
  destination_service_name = "<DESTINATION_SERVICE_NAME>"
  integration_type         = "datadog"
  source_service_name      = "<SOURCE_SERVICE_NAME>"
}

