resource "aiven_service_integration" "my_integration_metrics" {
  project                  = aiven_project.myproject.project
  integration_type         = "metrics"
  source_service_name      = aiven_kafka.kfk1.service_name
  destination_service_name = aiven_m3db.m3db.service_name
}
