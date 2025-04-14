resource "aiven_m3db" "m3db" {
  project      = data.aiven_project.m3db-project1.project
  cloud_name   = "google-europe-west1"
  plan         = "business-8"
  service_name = "m3db"

  m3db_user_config {
    m3db_version = 1.5

    namespaces {
      name = "m3db_ns"
      type = "unaggregated"
    }
  }
}

// Get PG data to M3
resource "aiven_pg" "pg1" {
  project      = data.aiven_project.m3db-project1.project
  cloud_name   = "google-europe-west1"
  service_name = "postgres1"
  plan         = "startup-4"

  pg_user_config {
    pg_version = 14
  }
}

resource "aiven_service_integration" "int-m3db-pg" {
  project                  = data.aiven_project.m3db-project1.project
  integration_type         = "metrics"
  source_service_name      = aiven_pg.pg1.service_name
  destination_service_name = aiven_m3db.m3db.service_name
}

// Grafana dashboard for M3
resource "aiven_grafana" "grafana1" {
  project      = data.aiven_project.m3db-project1.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "grafana1"

  grafana_user_config {
    alerting_enabled = true

    public_access {
      grafana = true
    }
  }
}

resource "aiven_service_integration" "int-grafana-m3db" {
  project                  = data.aiven_project.m3db-project1.project
  integration_type         = "dashboard"
  source_service_name      = aiven_grafana.grafana1.service_name
  destination_service_name = aiven_m3db.m3db.service_name
}


// Setting up aggregation
resource "aiven_m3aggregator" "m3a" {
  project      = data.aiven_project.m3db-project1.project
  cloud_name   = "google-europe-west1"
  plan         = "business-8"
  service_name = "m3a"

  m3aggregator_user_config {
    m3aggregator_version = 1.5
  }
}

resource "aiven_service_integration" "int-m3db-aggr" {
  project                  = data.aiven_project.m3db-project1.project
  integration_type         = "m3aggregator"
  source_service_name      = aiven_m3db.m3db.service_name
  destination_service_name = aiven_m3aggregator.m3a.service_name
}
