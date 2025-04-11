resource "aiven_opensearch" "example_opensearch" {
  project                 = data.aiven_project.example_project.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "example-opensearch"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  opensearch_user_config {

    opensearch_dashboards {
      enabled                    = true
      opensearch_request_timeout = 30000
    }

    public_access {
      opensearch            = true
      opensearch_dashboards = true
    }
  }
}
