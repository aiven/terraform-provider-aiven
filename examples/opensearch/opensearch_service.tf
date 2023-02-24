# Opensearch service
resource "aiven_opensearch" "os" {
  project                 = aiven_project.os-project.project
  cloud_name              = "google-europe-west1"
  plan                    = "hobbyist"
  service_name            = "os-service"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

# Opensearch user
resource "aiven_opensearch_user" "os-user" {
  project      = aiven_project.os-project.project
  service_name = aiven_opensearch.os.service_name
  username     = "test-user1"
}