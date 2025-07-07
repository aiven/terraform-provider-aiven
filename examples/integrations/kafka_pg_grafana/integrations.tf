# Integrate PostgreSQL with Grafana Metrics Dashboard
resource "aiven_service_integration" "pg_grafana_integration" {
  project                  = data.aiven_project.main.project
  integration_type         = "dashboard"
  source_service_name      = aiven_grafana.example_grafana.service_name
  destination_service_name = aiven_pg.example_pg.service_name
}

# Integrate Kafka with PostgreSQL to receive metrics from the Kafka service
resource "aiven_service_integration" "kafka_pg_integration" {
  project                  = data.aiven_project.main.project
  integration_type         = "metrics"
  source_service_name      = aiven_kafka.example_kafka.service_name
  destination_service_name = aiven_pg.example_pg.service_name
}
