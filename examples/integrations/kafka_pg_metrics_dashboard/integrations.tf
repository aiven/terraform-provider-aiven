# Send metrics from Kafka to Thanos
resource "aiven_service_integration" "samplekafka_metrics" {
  project                  = var.aiven_project_name
  integration_type         = "metrics"
  source_service_name      = aiven_kafka.samplekafka.service_name
  destination_service_name = aiven_thanos.example_thanos.service_name
}

# Send metrics from PostgreSQL to Thanos
resource "aiven_service_integration" "samplepg_metrics" {
  project                  = var.aiven_project_name
  integration_type         = "metrics"
  source_service_name      = aiven_pg.samplepg.service_name
  destination_service_name = aiven_thanos.example_thanos.service_name
}

# Dashboards for Kafka and PostgreSQL services
resource "aiven_service_integration" "samplegrafana_dashboards" {
  project                  = var.aiven_project_name
  integration_type         = "dashboard"
  source_service_name      = aiven_grafana.samplegrafana.service_name
  destination_service_name = aiven_thanos.example_thanos.service_name
}
