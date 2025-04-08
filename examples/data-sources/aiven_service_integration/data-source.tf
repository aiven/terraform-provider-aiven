data "aiven_service_integration" "example_integration" {
  project                  = data.aiven_project.example_project.project
  destination_service_name = aiven_thanos.example_thanos.service_name
  integration_type         = "metrics"
  source_service_name      = aiven_kafka.example_kafka.service_name
}

