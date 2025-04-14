# Your Aiven project
data "aiven_project" "main" {
  project = var.aiven_project_name
}

# OpenSearch service
resource "aiven_opensearch" "example_opensearch" {
  project                 = data.aiven_project.main.project
  cloud_name              = "google-europe-west1"
  plan                    = "hobbyist"
  service_name            = var.opensearch_service_name
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

# OpenSearch user
resource "aiven_opensearch_user" "os_user" {
  project      = data.aiven_project.main.project
  service_name = aiven_opensearch.example_opensearch.service_name
  username     = "example-opensearch-user"
}
