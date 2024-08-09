data "aiven_service_integration" "example_integration" {
  project                  = data.aiven_project.example_project.project
  destination_service_name = aiven_m3db.example_m3db.service_name
  integration_type         = "metrics"
  source_service_name      = aiven_kafka.example_kafka.service_name
}

