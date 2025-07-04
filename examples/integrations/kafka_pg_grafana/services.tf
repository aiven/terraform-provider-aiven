# Your Aiven project
data "aiven_project" "main" {
  project = var.aiven_project_name
}

# Kafka service
resource "aiven_kafka" "example_kafka" {
  project      = data.aiven_project.main.project
  service_name = var.kafka_service_name
  cloud_name   = "aws-eu-west-2"
  plan         = "startup-2"
}

# PostgreSQL service
resource "aiven_pg" "example_pg" {
  project      = data.aiven_project.main.project
  service_name = var.postgres_service_name
  cloud_name   = "aws-eu-west-2"
  plan         = "startup-4"
}

# Grafana service
resource "aiven_grafana" "example_grafana" {
  project      = data.aiven_project.main.project
  service_name = var.grafana_service_name
  cloud_name   = "aws-eu-west-2"
  plan         = "startup-4"

  grafana_user_config {
    alerting_enabled = true

    public_access {
      grafana = true
    }
  }
}
